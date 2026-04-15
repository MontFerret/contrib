package session

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
	. "github.com/smartystreets/goconvey/convey"
)

type targetAPI struct {
	cdp.Target

	attached          func(context.Context) (target.AttachedToTargetClient, error)
	detached          func(context.Context) (target.DetachedFromTargetClient, error)
	message           func(context.Context) (target.ReceivedMessageFromTargetClient, error)
	autoAttachRelated func(context.Context, *target.AutoAttachRelatedArgs) error
	detachFromTarget  func(context.Context, *target.DetachFromTargetArgs) error
	sendMessage       func(context.Context, *target.SendMessageToTargetArgs) error
}

func (api *targetAPI) AttachedToTarget(ctx context.Context) (target.AttachedToTargetClient, error) {
	return api.attached(ctx)
}

func (api *targetAPI) DetachedFromTarget(ctx context.Context) (target.DetachedFromTargetClient, error) {
	return api.detached(ctx)
}

func (api *targetAPI) ReceivedMessageFromTarget(ctx context.Context) (target.ReceivedMessageFromTargetClient, error) {
	return api.message(ctx)
}

func (api *targetAPI) AutoAttachRelated(ctx context.Context, args *target.AutoAttachRelatedArgs) error {
	if api.autoAttachRelated == nil {
		return nil
	}

	return api.autoAttachRelated(ctx, args)
}

func (api *targetAPI) DetachFromTarget(ctx context.Context, args *target.DetachFromTargetArgs) error {
	if api.detachFromTarget == nil {
		return nil
	}

	return api.detachFromTarget(ctx, args)
}

func (api *targetAPI) SendMessageToTarget(ctx context.Context, args *target.SendMessageToTargetArgs) error {
	if api.sendMessage == nil {
		return nil
	}

	return api.sendMessage(ctx, args)
}

type testStream struct {
	ready   chan struct{}
	message chan any
}

func newTestStream() *testStream {
	return &testStream{
		ready:   make(chan struct{}, 8),
		message: make(chan any, 8),
	}
}

func (s *testStream) Ready() <-chan struct{} {
	return s.ready
}

func (s *testStream) RecvMsg(any) error {
	return nil
}

func (s *testStream) Close() error {
	close(s.ready)
	close(s.message)
	return nil
}

func (s *testStream) emit(message any) {
	s.ready <- struct{}{}
	s.message <- message
}

type attachedClient struct{ *testStream }

func (c *attachedClient) Recv() (*target.AttachedToTargetReply, error) {
	reply, _ := (<-c.message).(*target.AttachedToTargetReply)
	return reply, nil
}

type detachedClient struct{ *testStream }

func (c *detachedClient) Recv() (*target.DetachedFromTargetReply, error) {
	reply, _ := (<-c.message).(*target.DetachedFromTargetReply)
	return reply, nil
}

type receivedMessageClient struct{ *testStream }

func (c *receivedMessageClient) Recv() (*target.ReceivedMessageFromTargetReply, error) {
	reply, _ := (<-c.message).(*target.ReceivedMessageFromTargetReply)
	return reply, nil
}

func TestManager(t *testing.T) {
	Convey("session manager bookkeeping", t, func() {
		oldAttachRoot := attachRootSessionClient
		oldAttachRelated := attachRelatedSessionClient
		oldEnable := enableRelatedSessionClient
		oldSync := syncEventStreams
		defer func() {
			attachRootSessionClient = oldAttachRoot
			attachRelatedSessionClient = oldAttachRelated
			enableRelatedSessionClient = oldEnable
			syncEventStreams = oldSync
		}()

		rootClosed := atomic.Int32{}
		childClosed := atomic.Int32{}

		root := &Client{
			ID:       "root-session",
			TargetID: "root-target",
			closeFn: func() error {
				rootClosed.Add(1)
				return nil
			},
		}

		child := &Client{
			ID:       "child-session",
			TargetID: "child-target",
			closeFn: func() error {
				childClosed.Add(1)
				return nil
			},
		}

		attachRootSessionClient = func(context.Context, *cdp.Client, target.ID) (*Client, error) {
			return root, nil
		}

		attachRelatedSessionClient = func(context.Context, *cdp.Client, *target.AttachedToTargetReply) (*Client, error) {
			return child, nil
		}

		enableRelatedSessionClient = func(context.Context, *cdp.Client) error {
			return nil
		}
		syncEventStreams = func(
			target.AttachedToTargetClient,
			target.DetachedFromTargetClient,
			target.ReceivedMessageFromTargetClient,
		) error {
			return nil
		}

		attached := &attachedClient{newTestStream()}
		detached := &detachedClient{newTestStream()}
		message := &receivedMessageClient{newTestStream()}

		client := &cdp.Client{
			Target: &targetAPI{
				attached: func(context.Context) (target.AttachedToTargetClient, error) {
					return attached, nil
				},
				detached: func(context.Context) (target.DetachedFromTargetClient, error) {
					return detached, nil
				},
				message: func(context.Context) (target.ReceivedMessageFromTargetClient, error) {
					return message, nil
				},
			},
		}

		manager, err := New(context.Background(), nil, client, "root-target")
		So(err, ShouldBeNil)
		defer func() {
			So(manager.Close(), ShouldBeNil)
		}()

		So(manager.Root(), ShouldEqual, root)
		So(len(manager.Snapshot()), ShouldEqual, 1)

		events := make(chan Event, 2)
		listenerID := manager.AddListener(func(event Event) {
			events <- event
		})
		defer manager.RemoveListener(listenerID)

		attached.emit(&target.AttachedToTargetReply{
			SessionID: "child-session",
			TargetInfo: target.Info{
				TargetID: "child-target",
			},
		})

		So(waitFor(func() bool {
			return len(manager.Snapshot()) == 2
		}), ShouldBeTrue)

		attachedEvent := <-events
		So(attachedEvent.Kind, ShouldEqual, EventAttached)
		So(attachedEvent.Client.TargetID, ShouldEqual, target.ID("child-target"))

		detached.emit(&target.DetachedFromTargetReply{
			SessionID: "child-session",
		})

		So(waitFor(func() bool {
			return len(manager.Snapshot()) == 1
		}), ShouldBeTrue)

		detachedEvent := <-events
		So(detachedEvent.Kind, ShouldEqual, EventDetached)
		So(detachedEvent.Client.TargetID, ShouldEqual, target.ID("child-target"))
		So(childClosed.Load(), ShouldEqual, 1)
	})
}

func TestClientCloseIgnoresMissingSession(t *testing.T) {
	api := &targetAPI{
		detachFromTarget: func(context.Context, *target.DetachFromTargetArgs) error {
			return errors.New("rpc error: No session with given id (code = -32602)")
		},
	}

	client := &cdp.Client{Target: api}

	sessionClient, err := newClientFromSession(
		context.Background(),
		client,
		target.Info{TargetID: "target-id"},
		"session-id",
	)
	if err != nil {
		t.Fatalf("newClientFromSession: %v", err)
	}

	if err := sessionClient.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func waitFor(predicate func() bool) bool {
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if predicate() {
			return true
		}

		time.Sleep(10 * time.Millisecond)
	}

	return predicate()
}
