package session

import (
	"context"

	"github.com/mafredri/cdp"
	protocoldom "github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
)

func enableAttachedClient(ctx context.Context, client *cdp.Client) error {
	if err := client.Page.Enable(ctx); err != nil {
		return err
	}

	return runBatch(
		func() error {
			return client.Page.SetLifecycleEventsEnabled(
				ctx,
				page.NewSetLifecycleEventsEnabledArgs(true),
			)
		},
		func() error {
			return client.DOM.Enable(ctx, protocoldom.NewEnableArgs().SetIncludeWhitespace("all"))
		},
		func() error {
			return client.Runtime.Enable(ctx)
		},
		func() error {
			return client.Network.Enable(ctx, network.NewEnableArgs())
		},
	)
}

func runBatch(funcs ...func() error) error {
	for _, fn := range funcs {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
