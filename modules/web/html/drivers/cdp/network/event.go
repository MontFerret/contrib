package network

import (
	"context"
	"strings"

	"github.com/goccy/go-json"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var NavigationEventType = runtime.NewTypeFor[*NavigationEvent]()

type (
	NavigationEvent struct {
		sourceClient *cdp.Client
		URL          string
		FrameID      page.FrameID
		MimeType     string
	}

	navigationEventJSON struct {
		URL      string `json:"url"`
		FrameID  string `json:"frame_id"`
		MimeType string `json:"mime_type"`
	}
)

func (evt *NavigationEvent) MarshalJSON() ([]byte, error) {
	if evt == nil {
		return json.Marshal(nil)
	}

	return json.Marshal(navigationEventJSON{
		URL:      evt.URL,
		FrameID:  string(evt.FrameID),
		MimeType: evt.MimeType,
	})
}

func (evt *NavigationEvent) Type() runtime.Type {
	return NavigationEventType
}

func (evt *NavigationEvent) String() string {
	return evt.URL
}

func (evt *NavigationEvent) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	otherEvt, ok := other.(*NavigationEvent)
	if !ok {
		return false, nil
	}

	comparison, err := evt.compare(ctx, otherEvt)

	return comparison == runtime.Equal, err
}

func (evt *NavigationEvent) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if err := ctx.Err(); err != nil {
		return runtime.Equal, err
	}

	otherEvt, ok := other.(*NavigationEvent)
	if !ok {
		return runtime.Equal, runtime.Errorf(
			runtime.ErrInvalidOperation,
			"cannot compare %s with %s",
			runtime.TypeName(runtime.TypeOf(evt)),
			runtime.TypeName(runtime.TypeOf(other)),
		)
	}

	return evt.compare(ctx, otherEvt)
}

func (evt *NavigationEvent) compare(ctx context.Context, other *NavigationEvent) (runtime.Ordering, error) {
	if err := ctx.Err(); err != nil {
		return runtime.Equal, err
	}

	if comparison := strings.Compare(evt.URL, other.URL); comparison != 0 {
		return runtime.Ordering(comparison), nil
	}

	return runtime.Ordering(strings.Compare(string(evt.FrameID), string(other.FrameID))), nil
}

func (evt *NavigationEvent) Unwrap() any {
	return evt
}

func (evt *NavigationEvent) Hash() uint64 {
	return runtime.Hash(
		runtime.TypeName(evt.Type()),
		[]byte(evt.URL+"\x00"+string(evt.FrameID)),
	)
}

func (evt *NavigationEvent) Copy() runtime.Value {
	return &NavigationEvent{
		URL:          evt.URL,
		FrameID:      evt.FrameID,
		MimeType:     evt.MimeType,
		sourceClient: evt.sourceClient,
	}
}

func (evt *NavigationEvent) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	switch key.String() {
	case "url", "URL":
		return runtime.NewString(evt.URL), nil
	case "frame":
		return dom.NewFrameID(evt.FrameID), nil
	default:
		return runtime.None, nil
	}
}

func (evt *NavigationEvent) SourceClient() *cdp.Client {
	if evt == nil {
		return nil
	}

	return evt.sourceClient
}
