package network

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestNavigateRetainsEventBeforeCommandReturns(t *testing.T) {
	api := newNavigationTestAPI()
	manager := navigationTestManager(api)
	api.navigate = func() (*page.NavigateReply, error) {
		if err := emitImmediateNavigation(api); err != nil {
			return nil, err
		}

		return &page.NavigateReply{FrameID: "root"}, nil
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := manager.Navigate(ctx, "https://example.test/next"); err != nil {
		t.Fatal(err)
	}
}

func TestHistoryNavigationRetainsImmediateEvent(t *testing.T) {
	for _, direction := range []string{"back", "forward"} {
		t.Run(direction, func(t *testing.T) {
			api := newNavigationTestAPI()
			api.history = &page.GetNavigationHistoryReply{
				Entries: []page.NavigationEntry{{ID: 1}, {ID: 2}},
			}

			if direction == "back" {
				api.history.CurrentIndex = 1
			}

			api.change = func() error { return emitImmediateNavigation(api) }
			manager := navigationTestManager(api)
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()

			var moved runtime.Boolean
			var err error
			if direction == "back" {
				moved, err = manager.NavigateBack(ctx, 1)
			} else {
				moved, err = manager.NavigateForward(ctx, 1)
			}

			if err != nil || !moved {
				t.Fatalf("moved=%v error=%v", moved, err)
			}
		})
	}
}

func TestHistoryNavigationNoOpDoesNotSubscribe(t *testing.T) {
	for _, direction := range []string{"back", "forward"} {
		t.Run(direction, func(t *testing.T) {
			api := newNavigationTestAPI()
			api.history = &page.GetNavigationHistoryReply{Entries: []page.NavigationEntry{{ID: 1}}}
			manager := navigationTestManager(api)
			var moved runtime.Boolean
			var err error
			if direction == "back" {
				moved, err = manager.NavigateBack(t.Context(), 1)
			} else {
				moved, err = manager.NavigateForward(t.Context(), 1)
			}

			if err != nil || moved || api.subscriptions.Load() != 0 || api.commands.Load() != 0 {
				t.Fatalf("no-op moved=%v error=%v subscriptions=%d commands=%d", moved, err, api.subscriptions.Load(), api.commands.Load())
			}
		})
	}
}

func TestNavigationSetupFailurePreventsCommand(t *testing.T) {
	for _, phase := range []string{"frame", "within document"} {
		t.Run(phase, func(t *testing.T) {
			api := newNavigationTestAPI()
			want := errors.New("subscription setup failed")
			if phase == "frame" {
				api.frameErr = want
			} else {
				api.withinErr = want
			}

			manager := navigationTestManager(api)
			err := manager.Navigate(t.Context(), "https://example.test/")
			if !errors.Is(err, want) || api.commands.Load() != 0 {
				t.Fatalf("error=%v commands=%d", err, api.commands.Load())
			}

			if manager.sessions.(*subscriptionTestSessions).listenerCount() != 0 {
				t.Fatal("failed setup retained a session listener")
			}
		})
	}
}

func TestNavigationFailureClosesSubscription(t *testing.T) {
	api := newNavigationTestAPI()
	want := errors.New("navigation failed")
	api.navigate = func() (*page.NavigateReply, error) { return nil, want }
	manager := navigationTestManager(api)
	if err := manager.Navigate(t.Context(), "https://example.test/"); !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}

	if manager.sessions.(*subscriptionTestSessions).listenerCount() != 0 {
		t.Fatal("failed navigation retained a session listener")
	}
}

func navigationTestManager(api *navigationTestAPI) *Manager {
	client := &cdp.Client{Page: api}

	return &Manager{
		logger:   zerolog.Nop(),
		client:   client,
		sessions: newSubscriptionTestSessions(&cdpsession.Client{ID: "root", CDP: client}),
	}
}

func emitImmediateNavigation(api *navigationTestAPI) error {
	if api.subscriptions.Load() == 0 {
		return errors.New("navigation started before subscription")
	}

	api.frame.reply = &page.FrameNavigatedReply{Frame: page.Frame{ID: "root", URL: "https://example.test/next"}}
	close(api.frame.ready)

	return nil
}
