package network

import (
	"context"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

type Interceptor struct {
	logger   zerolog.Logger
	client   *cdp.Client
	filters  map[string]*InterceptorFilter
	loop     *events.Loop
	mu       sync.RWMutex
	running  bool
	cancel   context.CancelFunc
	done     chan struct{}
	err      error
	requests sync.WaitGroup
}

type InterceptorListener func(ctx context.Context, msg *fetch.RequestPausedReply) bool

func NewInterceptor(logger zerolog.Logger, client *cdp.Client) *Interceptor {
	i := new(Interceptor)
	i.logger = logutil.WithComponent(logger.With(), "network_interceptor").Logger()
	i.client = client
	i.filters = make(map[string]*InterceptorFilter)
	i.loop = events.NewLoop(createRequestPausedStreamFactory(client))
	i.loop.AddListener(requestPausedEvent, events.Always(i.filter))
	i.loop.AddListener(events.Error, events.Always(i.failed))
	i.done = make(chan struct{})

	return i
}

func (i *Interceptor) IsRunning() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.running
}

func (i *Interceptor) AddFilter(name string, filter *Filter) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	f, err := NewInterceptorFilter(filter)
	if err != nil {
		return err
	}

	i.filters[name] = f

	return nil
}

func (i *Interceptor) RemoveFilter(name string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.filters, name)
}

func (i *Interceptor) AddListener(listener InterceptorListener) events.ListenerID {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.loop.AddListener(requestPausedEvent, func(ctx context.Context, message any) bool {
		msg, ok := message.(*fetch.RequestPausedReply)

		if !ok {
			return true
		}

		return listener(ctx, msg)
	})
}

func (i *Interceptor) RemoveListener(id events.ListenerID) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.loop.RemoveListener(requestPausedEvent, id)
}

// Run subscribes before enabling Fetch so no paused request can be lost at startup.
// The context owns interception until cancellation. A stream failure leaves Fetch
// enabled, keeping requests paused until the owner tears down the connection.
func (i *Interceptor) Run(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.running {
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	i.cancel = cancel
	i.err = nil
	i.done = make(chan struct{})

	if err := i.loop.Run(runCtx); err != nil {
		cancel()
		i.err = err
		close(i.done)

		return err
	}

	if err := i.client.Fetch.Enable(runCtx, fetch.NewEnableArgs()); err != nil {
		cancel()
		i.err = err
		close(i.done)

		return err
	}

	i.running = true
	go i.waitForStop(runCtx)

	return nil
}

// Done closes after interception stops. Err distinguishes stream failure from
// normal cancellation. Callers must finish Run before observing Done.
func (i *Interceptor) Done() <-chan struct{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.done
}

// Err reports an unexpected interception stream failure or a startup error.
func (i *Interceptor) Err() error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.err
}

// Close cancels interception and waits for bounded Fetch cleanup.
func (i *Interceptor) Close() error {
	i.mu.RLock()
	cancel, done := i.cancel, i.done
	i.mu.RUnlock()

	if cancel == nil {
		return nil
	}

	cancel()
	<-done

	return i.Err()
}

func (i *Interceptor) failed(_ context.Context, message any) {
	err, ok := message.(error)
	if !ok {
		return
	}

	i.mu.Lock()
	if i.err == nil {
		i.err = err
	}

	cancel := i.cancel
	i.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (i *Interceptor) waitForStop(ctx context.Context) {
	<-ctx.Done()

	i.mu.Lock()
	failed := i.err != nil
	i.mu.Unlock()
	i.requests.Wait()

	if !failed {
		nested, cancel := context.WithTimeout(context.Background(), time.Duration(drivers.DefaultWaitTimeout)*time.Millisecond)
		defer cancel()

		if err := i.client.Fetch.Disable(nested); err != nil {
			i.logger.Trace().Err(err).Msg("failed to stop interceptor")
		}
	}

	i.mu.Lock()
	i.running = false
	close(i.done)
	i.mu.Unlock()
}

func (i *Interceptor) filter(ctx context.Context, message any) {
	msg, ok := message.(*fetch.RequestPausedReply)
	if !ok {
		return
	}

	log := i.logger.With().
		Str("request_id", string(msg.RequestID)).
		Str("frame_id", string(msg.FrameID)).
		Str("resource_type", string(msg.ResourceType)).
		Str("url", msg.Request.URL).
		Logger()

	log.Trace().Msg("trying to block resource loading")

	var reject bool

	i.mu.Lock()
	if ctx.Err() != nil {
		i.mu.Unlock()

		return
	}

	for _, filter := range i.filters {
		reject = filter.Filter(msg.ResourceType, msg.Request)

		if reject {
			break
		}
	}

	i.requests.Add(1)
	i.mu.Unlock()

	defer i.requests.Done()
	i.resolve(ctx, msg, reject)
}

func (i *Interceptor) resolve(ctx context.Context, msg *fetch.RequestPausedReply, reject bool) {
	if !reject {
		err := i.client.Fetch.ContinueRequest(ctx, fetch.NewContinueRequestArgs(msg.RequestID))
		if err != nil {
			i.logger.Err(err).Msg("failed to allow resource loading")
		}

		return
	}

	err := i.client.Fetch.FailRequest(ctx, fetch.NewFailRequestArgs(msg.RequestID, network.ErrorReasonBlockedByClient))
	if err != nil {
		i.logger.Trace().Err(err).Msg("failed to block resource loading")
	}
}
