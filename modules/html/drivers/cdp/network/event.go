package network

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/goccy/go-json"
	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/dom"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var NavigationEventType = runtime.NewTypeFor[*NavigationEvent]("html.drivers.cdp.network", "NavigationEvent")

type (
	NavigationEvent struct {
		URL      string
		FrameID  page.FrameID
		MimeType string
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

func (evt *NavigationEvent) Compare(other runtime.Value) int {
	otherEvt, ok := other.(*NavigationEvent)

	if !ok {
		return drivers.CompareTypes(evt, other)
	}

	comp := runtime.NewString(evt.URL).Compare(runtime.NewString(otherEvt.URL))

	if comp != 0 {
		return comp
	}

	return runtime.String(evt.FrameID).Compare(runtime.String(otherEvt.FrameID))
}

func (evt *NavigationEvent) Unwrap() any {
	return evt
}

func (evt *NavigationEvent) Hash() uint64 {
	return runtime.Parse(evt).Hash()
}

func (evt *NavigationEvent) Copy() runtime.Value {
	return *(&evt)
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
