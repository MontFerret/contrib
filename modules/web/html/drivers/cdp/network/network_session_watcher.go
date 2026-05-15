package network

import (
	"context"
	"errors"
	"sync"

	"github.com/mafredri/cdp"
	cdpnetwork "github.com/mafredri/cdp/protocol/network"
	"github.com/rs/zerolog"
)

type networkSessionWatcher struct {
	logger    zerolog.Logger
	ctx       context.Context
	request   cdpnetwork.RequestWillBeSentClient
	response  cdpnetwork.ResponseReceivedClient
	finished  cdpnetwork.LoadingFinishedClient
	failed    cdpnetwork.LoadingFailedClient
	fromCache cdpnetwork.RequestServedFromCacheClient
	closeErr  error
	client    *cdp.Client
	cancel    context.CancelFunc
	key       string
	closeOnce sync.Once
}

func newNetworkSessionWatcher(
	ctx context.Context,
	logger zerolog.Logger,
	key string,
	client *cdp.Client,
) (*networkSessionWatcher, error) {
	if client == nil || client.Network == nil {
		return nil, nil
	}

	watcherCtx, cancel := context.WithCancel(ctx)
	watcher := &networkSessionWatcher{
		logger: logger,
		client: client,
		ctx:    watcherCtx,
		cancel: cancel,
		key:    key,
	}

	var err error

	watcher.request, err = client.Network.RequestWillBeSent(watcherCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	watcher.response, err = client.Network.ResponseReceived(watcherCtx)
	if err != nil {
		_ = watcher.Close()
		return nil, err
	}

	watcher.finished, err = client.Network.LoadingFinished(watcherCtx)
	if err != nil {
		_ = watcher.Close()
		return nil, err
	}

	watcher.failed, err = client.Network.LoadingFailed(watcherCtx)
	if err != nil {
		_ = watcher.Close()
		return nil, err
	}

	watcher.fromCache, err = client.Network.RequestServedFromCache(watcherCtx)
	if err != nil {
		_ = watcher.Close()
		return nil, err
	}

	return watcher, nil
}

func (w *networkSessionWatcher) Close() error {
	if w == nil {
		return nil
	}

	w.closeOnce.Do(func() {
		if w.cancel != nil {
			w.cancel()
		}

		w.closeErr = errors.Join(
			closeNetworkStream(w.request),
			closeNetworkStream(w.response),
			closeNetworkStream(w.finished),
			closeNetworkStream(w.failed),
			closeNetworkStream(w.fromCache),
		)
	})

	return w.closeErr
}

func (w *networkSessionWatcher) Run(observer *networkObserver) {
	defer w.Close()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.request.Ready():
			if w.ctx.Err() != nil {
				return
			}

			reply, err := w.request.Recv()
			if err != nil {
				w.emitError(observer, err, "failed to receive request started event")
				return
			}

			observer.handleRequestStarted(w.key, w.client, reply)
		case <-w.response.Ready():
			if w.ctx.Err() != nil {
				return
			}

			reply, err := w.response.Recv()
			if err != nil {
				w.emitError(observer, err, "failed to receive response event")
				return
			}

			observer.handleResponseReceived(w.key, w.client, reply)
		case <-w.finished.Ready():
			if w.ctx.Err() != nil {
				return
			}

			reply, err := w.finished.Recv()
			if err != nil {
				w.emitError(observer, err, "failed to receive request finished event")
				return
			}

			observer.handleRequestFinished(w.key, w.client, reply)
		case <-w.failed.Ready():
			if w.ctx.Err() != nil {
				return
			}

			reply, err := w.failed.Recv()
			if err != nil {
				w.emitError(observer, err, "failed to receive request failed event")
				return
			}

			observer.handleRequestFailed(w.key, w.client, reply)
		case <-w.fromCache.Ready():
			if w.ctx.Err() != nil {
				return
			}

			reply, err := w.fromCache.Recv()
			if err != nil {
				w.emitError(observer, err, "failed to receive request served from cache event")
				return
			}

			observer.handleRequestServedFromCache(w.key, reply)
		}
	}
}

func (w *networkSessionWatcher) emitError(observer *networkObserver, err error, message string) {
	if w.ctx.Err() != nil {
		return
	}

	w.logger.Trace().Err(err).Msg(message)
	observer.emit(networkEvent{err: err})
}
