package network_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
)

func TestManagerOriginFilters(t *testing.T) {
	for _, test := range []struct {
		name   string
		filter *network.Filter
		enable bool
		fail   bool
	}{
		{name: "omitted"},
		{name: "empty filter", filter: &network.Filter{}},
		{name: "deny all", filter: &network.Filter{AllowedOrigins: []string{}}, enable: true},
		{name: "origin only", filter: &network.Filter{AllowedOrigins: []string{"https://example.com"}}, enable: true},
		{name: "invalid origin", filter: &network.Filter{AllowedOrigins: []string{"invalid"}}, fail: true},
		{name: "resources", filter: &network.Filter{Patterns: []drivers.ResourceFilter{{Type: "Image", URL: "*.png"}}}, enable: true},
		{name: "combined", filter: &network.Filter{Patterns: []drivers.ResourceFilter{{Type: "Image", URL: "*.png"}}, AllowedOrigins: []string{"https://example.com"}}, enable: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var enables, disables atomic.Int32
			api := &FetchAPI{
				enable: func(context.Context, *fetch.EnableArgs) error {
					enables.Add(1)

					return nil
				},
				disable: func(context.Context) error {
					disables.Add(1)

					return nil
				},
				requestPaused: func(context.Context) (fetch.RequestPausedClient, error) { return NewRequestPausedClient(), nil },
			}
			manager, err := network.New(zerolog.Nop(), &cdp.Client{Network: &NetworkAPI{}, Fetch: api}, nil, network.Options{Filter: test.filter})
			if (err != nil) != test.fail {
				t.Fatalf("New error = %v", err)
			}

			if manager != nil {
				_ = manager.Close()
			}

			want := int32(0)
			if test.enable {
				want = 1
			}

			if enables.Load() != want || disables.Load() != want {
				t.Fatalf("Fetch enable=%d disable=%d, want %d", enables.Load(), disables.Load(), want)
			}
		})
	}
}
