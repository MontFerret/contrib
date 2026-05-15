package network

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/mafredri/cdp"
	cdpnetwork "github.com/mafredri/cdp/protocol/network"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type networkObserver struct {
	logger           zerolog.Logger
	client           *cdp.Client
	sessions         *cdpsession.Manager
	ctx              context.Context
	cancel           context.CancelFunc
	subscribers      map[int64]*networkEventSubscriber
	watchers         map[string]*networkSessionWatcher
	requests         map[string]networkRequestState
	listenerID       cdpsession.ListenerID
	nextSubscriberID atomic.Int64
	wg               sync.WaitGroup
	mu               sync.Mutex
	closeOnce        sync.Once
}

func newNetworkObserver(logger zerolog.Logger, client *cdp.Client, sessions *cdpsession.Manager) *networkObserver {
	return &networkObserver{
		logger:      logger,
		client:      client,
		sessions:    sessions,
		subscribers: make(map[int64]*networkEventSubscriber),
		watchers:    make(map[string]*networkSessionWatcher),
		requests:    make(map[string]networkRequestState),
	}
}

func (o *networkObserver) Start(ctx context.Context) error {
	if o == nil {
		return nil
	}

	o.ctx, o.cancel = context.WithCancel(ctx)

	if o.sessions == nil {
		return o.watchClient(rootSessionKey, o.client)
	}

	for _, client := range o.sessions.Snapshot() {
		if err := o.watchSession(client); err != nil {
			_ = o.Close()
			return err
		}
	}

	o.listenerID = o.sessions.AddListener(func(event cdpsession.Event) {
		switch event.Kind {
		case cdpsession.EventAttached:
			if err := o.watchSession(event.Client); err != nil {
				o.emit(networkEvent{err: err})
				o.logger.Warn().Err(err).Msg("failed to watch network stream for attached session")
			}
		case cdpsession.EventDetached:
			o.detachSession(event.Client)
		}
	})

	return nil
}

func (o *networkObserver) Close() error {
	if o == nil {
		return nil
	}

	o.closeOnce.Do(func() {
		if o.cancel != nil {
			o.cancel()
		}

		if o.sessions != nil {
			o.sessions.RemoveListener(o.listenerID)
		}

		o.closeWatchers()
		o.wg.Wait()
	})

	return nil
}

func (o *networkObserver) Subscribe(
	ctx context.Context,
	eventName string,
	options runtime.Map,
) (runtime.Stream, error) {
	switch eventName {
	case drivers.NetworkRequestStartedEvent,
		drivers.NetworkResponseReceivedEvent,
		drivers.NetworkRequestFinishedEvent,
		drivers.NetworkRequestFailedEvent:
		eventOptions, err := parseNetworkEventOptions(ctx, eventName, options)
		if err != nil {
			return nil, err
		}

		return newNetworkEventStream(o, o.logger, eventName, eventOptions), nil
	case drivers.NetworkIdleEvent:
		idleOptions, err := parseNetworkIdleOptions(ctx, eventName, options)
		if err != nil {
			return nil, err
		}

		return newNetworkIdleStream(o, idleOptions), nil
	default:
		return nil, invalidNetworkEventNameError(eventName)
	}
}

func (o *networkObserver) subscribe() *networkEventSubscriber {
	id := o.nextSubscriberID.Add(1)
	subscriber := &networkEventSubscriber{
		id:   id,
		ch:   make(chan networkEvent, 128),
		done: make(chan struct{}),
	}

	o.mu.Lock()
	o.subscribers[id] = subscriber
	o.mu.Unlock()

	return subscriber
}

func (o *networkObserver) unsubscribe(id int64) {
	o.mu.Lock()
	subscriber, exists := o.subscribers[id]
	if exists {
		delete(o.subscribers, id)
	}
	o.mu.Unlock()

	if exists {
		close(subscriber.done)
	}
}

func (o *networkObserver) emit(event networkEvent) {
	o.mu.Lock()
	subscribers := make([]*networkEventSubscriber, 0, len(o.subscribers))
	for _, subscriber := range o.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	o.mu.Unlock()

	for _, subscriber := range subscribers {
		select {
		case <-o.ctx.Done():
			return
		case <-subscriber.done:
		case subscriber.ch <- event:
		}
	}
}

func (o *networkObserver) snapshotActive(types map[string]struct{}) map[string]networkEvent {
	o.mu.Lock()
	defer o.mu.Unlock()

	active := make(map[string]networkEvent)
	for key, state := range o.requests {
		if len(types) > 0 {
			if _, exists := types[state.resourceType]; !exists {
				continue
			}
		}

		active[key] = networkEventFromState(drivers.NetworkRequestStartedEvent, state)
	}

	return active
}

func (o *networkObserver) watchSession(client *cdpsession.Client) error {
	if client == nil || client.CDP == nil {
		return nil
	}

	return o.watchClient(string(client.ID), client.CDP)
}

func (o *networkObserver) watchClient(key string, client *cdp.Client) error {
	if client == nil || client.Network == nil {
		return nil
	}

	o.mu.Lock()
	if _, exists := o.watchers[key]; exists {
		o.mu.Unlock()
		return nil
	}
	o.mu.Unlock()

	watcher, err := newNetworkSessionWatcher(o.ctx, o.logger, key, client)
	if err != nil {
		return err
	}

	if watcher == nil {
		return nil
	}

	o.mu.Lock()
	o.watchers[key] = watcher
	o.mu.Unlock()

	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		defer o.removeWatcher(key)
		watcher.Run(o)
	}()

	return nil
}

func (o *networkObserver) detachSession(client *cdpsession.Client) {
	if client == nil {
		return
	}

	key := string(client.ID)
	o.closeWatcher(key)
	o.handleSessionDetached(key)
}

func (o *networkObserver) closeWatcher(key string) {
	o.mu.Lock()
	watcher, exists := o.watchers[key]
	if exists {
		delete(o.watchers, key)
	}
	o.mu.Unlock()

	if exists {
		_ = watcher.Close()
	}
}

func (o *networkObserver) closeWatchers() {
	o.mu.Lock()
	watchers := make([]*networkSessionWatcher, 0, len(o.watchers))
	for key, watcher := range o.watchers {
		watchers = append(watchers, watcher)
		delete(o.watchers, key)
	}
	o.mu.Unlock()

	for _, watcher := range watchers {
		_ = watcher.Close()
	}
}

func (o *networkObserver) removeWatcher(key string) {
	o.mu.Lock()
	delete(o.watchers, key)
	o.mu.Unlock()
}

func (o *networkObserver) handleRequestStarted(
	sessionKey string,
	client *cdp.Client,
	reply *cdpnetwork.RequestWillBeSentReply,
) {
	if reply == nil {
		return
	}

	state := networkRequestState{
		sessionKey:     sessionKey,
		requestID:      reply.RequestID,
		loaderID:       reply.LoaderID,
		frameID:        frameIDString(reply.FrameID),
		url:            reply.Request.URL,
		method:         reply.Request.Method,
		resourceType:   normalizeResourceType(reply.Type),
		requestHeaders: toDriverHeaders(reply.Request.Headers),
		timestamp:      float64(reply.Timestamp),
		wallTime:       float64(reply.WallTime),
		client:         client,
	}

	key := networkRequestKey(sessionKey, reply.RequestID)

	o.mu.Lock()
	o.requests[key] = state
	o.mu.Unlock()

	o.emit(networkEventFromState(drivers.NetworkRequestStartedEvent, state))
}

func (o *networkObserver) handleResponseReceived(
	sessionKey string,
	client *cdp.Client,
	reply *cdpnetwork.ResponseReceivedReply,
) {
	if reply == nil {
		return
	}

	key := networkRequestKey(sessionKey, reply.RequestID)

	o.mu.Lock()
	state := o.requests[key]
	state.sessionKey = sessionKey
	state.requestID = reply.RequestID
	state.loaderID = reply.LoaderID
	state.frameID = frameIDString(reply.FrameID)
	state.url = reply.Response.URL
	if resourceType := normalizeResourceType(reply.Type); resourceType != "" {
		state.resourceType = resourceType
	}
	state.status = reply.Response.Status
	state.statusText = reply.Response.StatusText
	state.mimeType = reply.Response.MimeType
	state.headers = toDriverHeaders(reply.Response.Headers)
	if len(reply.Response.RequestHeaders) > 0 {
		state.requestHeaders = toDriverHeaders(reply.Response.RequestHeaders)
	}
	state.fromDiskCache = boolPtrValue(reply.Response.FromDiskCache)
	state.fromServiceWorker = boolPtrValue(reply.Response.FromServiceWorker)
	state.fromPrefetchCache = boolPtrValue(reply.Response.FromPrefetchCache)
	state.fromCache = state.fromCache || state.fromDiskCache || state.fromServiceWorker || state.fromPrefetchCache
	state.encodedDataLength = reply.Response.EncodedDataLength
	state.timestamp = float64(reply.Timestamp)
	state.client = client
	o.requests[key] = state
	o.mu.Unlock()

	o.emit(networkEventFromState(drivers.NetworkResponseReceivedEvent, state))
}

func (o *networkObserver) handleRequestFinished(
	sessionKey string,
	client *cdp.Client,
	reply *cdpnetwork.LoadingFinishedReply,
) {
	if reply == nil {
		return
	}

	key := networkRequestKey(sessionKey, reply.RequestID)

	o.mu.Lock()
	state := o.requests[key]
	delete(o.requests, key)
	o.mu.Unlock()

	state.sessionKey = sessionKey
	state.requestID = reply.RequestID
	state.encodedDataLength = reply.EncodedDataLength
	state.timestamp = float64(reply.Timestamp)
	if state.client == nil {
		state.client = client
	}

	o.emit(networkEventFromState(drivers.NetworkRequestFinishedEvent, state))
}

func (o *networkObserver) handleRequestFailed(
	sessionKey string,
	client *cdp.Client,
	reply *cdpnetwork.LoadingFailedReply,
) {
	if reply == nil {
		return
	}

	key := networkRequestKey(sessionKey, reply.RequestID)

	o.mu.Lock()
	state := o.requests[key]
	delete(o.requests, key)
	o.mu.Unlock()

	state.sessionKey = sessionKey
	state.requestID = reply.RequestID
	if resourceType := normalizeResourceType(reply.Type); resourceType != "" {
		state.resourceType = resourceType
	}
	state.failed = true
	state.errorText = reply.ErrorText
	state.canceled = boolPtrValue(reply.Canceled)
	state.blockedReason = reply.BlockedReason.String()
	state.timestamp = float64(reply.Timestamp)
	if state.client == nil {
		state.client = client
	}

	o.emit(networkEventFromState(drivers.NetworkRequestFailedEvent, state))
}

func (o *networkObserver) handleRequestServedFromCache(
	sessionKey string,
	reply *cdpnetwork.RequestServedFromCacheReply,
) {
	if reply == nil {
		return
	}

	key := networkRequestKey(sessionKey, reply.RequestID)

	o.mu.Lock()
	state := o.requests[key]
	state.sessionKey = sessionKey
	state.requestID = reply.RequestID
	state.fromCache = true
	o.requests[key] = state
	o.mu.Unlock()
}

func (o *networkObserver) handleSessionDetached(sessionKey string) {
	o.mu.Lock()
	for key := range o.requests {
		if strings.HasPrefix(key, sessionKey+"\x00") {
			delete(o.requests, key)
		}
	}
	o.mu.Unlock()

	o.emit(networkEvent{
		name:       networkSessionDetachedEvent,
		sessionKey: sessionKey,
	})
}
