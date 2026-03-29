package drivers

import (
	"strings"

	"github.com/goccy/go-json"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ScrollBehavior defines the transition animation.
// In HTML specification, default Value is auto, but in Ferret it's instant.
// More details here https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView
type ScrollBehavior string

const (
	ScrollBehaviorInstant ScrollBehavior = "instant"
	ScrollBehaviorSmooth  ScrollBehavior = "smooth"
	ScrollBehaviorAuto    ScrollBehavior = "auto"
)

func NewScrollBehavior(value string) ScrollBehavior {
	switch strings.ToLower(value) {
	case "instant":
		return ScrollBehaviorInstant
	case "smooth":
		return ScrollBehaviorSmooth
	default:
		return ScrollBehaviorAuto
	}
}

func (b ScrollBehavior) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

func (b ScrollBehavior) String() string {
	switch b {
	case ScrollBehaviorInstant:
		return "instant"
	case ScrollBehaviorSmooth:
		return "smooth"
	default:
		return "auto"
	}
}

// ScrollVerticalAlignment defines vertical alignment after scrolling.
// In HTML specification, default Value is start, but in Ferret it's center.
// More details here https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView
type ScrollVerticalAlignment string

const (
	ScrollVerticalAlignmentCenter  ScrollVerticalAlignment = "center"
	ScrollVerticalAlignmentStart   ScrollVerticalAlignment = "start"
	ScrollVerticalAlignmentEnd     ScrollVerticalAlignment = "end"
	ScrollVerticalAlignmentNearest ScrollVerticalAlignment = "nearest"
)

func NewScrollVerticalAlignment(value string) ScrollVerticalAlignment {
	switch strings.ToLower(value) {
	case "center":
		return ScrollVerticalAlignmentCenter
	case "start":
		return ScrollVerticalAlignmentStart
	case "end":
		return ScrollVerticalAlignmentEnd
	case "nearest":
		return ScrollVerticalAlignmentNearest
	default:
		return ScrollVerticalAlignmentCenter
	}
}

func (a ScrollVerticalAlignment) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a ScrollVerticalAlignment) String() string {
	switch a {
	case ScrollVerticalAlignmentCenter:
		return "center"
	case ScrollVerticalAlignmentStart:
		return "start"
	case ScrollVerticalAlignmentEnd:
		return "end"
	case ScrollVerticalAlignmentNearest:
		return "nearest"
	default:
		return "center"
	}
}

// ScrollHorizontalAlignment defines horizontal alignment after scrolling.
// In HTML specification, default Value is nearest, but in Ferret it's center.
// More details here https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView
type ScrollHorizontalAlignment string

const (
	ScrollHorizontalAlignmentCenter  ScrollHorizontalAlignment = "center"
	ScrollHorizontalAlignmentStart   ScrollHorizontalAlignment = "start"
	ScrollHorizontalAlignmentEnd     ScrollHorizontalAlignment = "end"
	ScrollHorizontalAlignmentNearest ScrollHorizontalAlignment = "nearest"
)

func NewScrollHorizontalAlignment(value string) ScrollHorizontalAlignment {
	switch strings.ToLower(value) {
	case "center":
		return ScrollHorizontalAlignmentCenter
	case "start":
		return ScrollHorizontalAlignmentStart
	case "end":
		return ScrollHorizontalAlignmentEnd
	case "nearest":
		return ScrollHorizontalAlignmentNearest
	default:
		return ScrollHorizontalAlignmentCenter
	}
}

func (a ScrollHorizontalAlignment) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a ScrollHorizontalAlignment) String() string {
	switch a {
	case ScrollHorizontalAlignmentCenter:
		return "center"
	case ScrollHorizontalAlignmentNearest:
		return "nearest"
	case ScrollHorizontalAlignmentStart:
		return "start"
	case ScrollHorizontalAlignmentEnd:
		return "end"
	default:
		return "center"
	}
}

// ScrollOptions defines how scroll animation should be performed.
type ScrollOptions struct {
	Behavior ScrollBehavior            `json:"behavior"`
	Block    ScrollVerticalAlignment   `json:"block"`
	Inline   ScrollHorizontalAlignment `json:"inline"`
	Top      runtime.Float             `json:"top"`
	Left     runtime.Float             `json:"left"`
}
