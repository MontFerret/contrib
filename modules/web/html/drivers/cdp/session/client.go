package session

import (
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

type Client struct {
	popup          bool
	runtimeOnly    bool
	Info           target.Info
	Conn           *rpcc.Conn
	CDP            *cdp.Client
	closeFn        func() error
	markDetachedFn func()
	writeFn        func([]byte) error
	ID             target.SessionID
	TargetID       target.ID
}

func (c *Client) Close() error {
	if c == nil || c.closeFn == nil {
		return nil
	}

	return c.closeFn()
}

func (c *Client) markDetached() {
	if c == nil || c.markDetachedFn == nil {
		return
	}

	c.markDetachedFn()
}

func (c *Client) writeMessage(message []byte) error {
	if c == nil || c.writeFn == nil {
		return nil
	}

	return c.writeFn(message)
}

// BelongsToPopup reports whether this session belongs to a separate popup document hierarchy.
func (c *Client) BelongsToPopup() bool { return c.popup }

// SupportsNetwork reports whether the target provides CDP Network events.
func (c *Client) SupportsNetwork() bool { return !c.runtimeOnly }
