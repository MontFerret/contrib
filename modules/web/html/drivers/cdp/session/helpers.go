package session

import (
	"context"
	"strings"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
)

func attachRootClient(
	ctx context.Context,
	browserClient *cdp.Client,
	targetID target.ID,
) (*Client, error) {
	args := target.NewAttachToTargetArgs(targetID)
	args.SetFlatten(false)

	reply, err := browserClient.Target.AttachToTarget(ctx, args)
	if err != nil {
		return nil, err
	}

	info := target.Info{TargetID: targetID}

	return newClientFromSession(ctx, browserClient, info, reply.SessionID)
}

func attachRelatedClient(
	ctx context.Context,
	browserClient *cdp.Client,
	reply *target.AttachedToTargetReply,
) (*Client, error) {
	if reply == nil {
		return nil, nil
	}

	return newClientFromSession(ctx, browserClient, reply.TargetInfo, reply.SessionID)
}

func newClientFromSession(
	ctx context.Context,
	browserClient *cdp.Client,
	info target.Info,
	sessionID target.SessionID,
) (*Client, error) {
	transport, err := newTransport(ctx, browserClient, info.TargetID, sessionID)
	if err != nil {
		return nil, err
	}

	conn := transport.Conn()

	return &Client{
		ID:             sessionID,
		TargetID:       info.TargetID,
		Info:           info,
		Conn:           conn,
		CDP:            cdp.NewClient(conn),
		closeFn:        transport.Close,
		markDetachedFn: transport.MarkDetached,
		writeFn:        transport.Write,
	}, nil
}

func isMissingSessionError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "No session with given id")
}
