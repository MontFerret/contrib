package dom

import (
	"testing"

	"github.com/mafredri/cdp/protocol/page"
)

func TestCollectFrameIDs(t *testing.T) {
	t.Parallel()

	tree := page.FrameTree{
		Frame: page.Frame{
			ID: "root",
		},
		ChildFrames: []page.FrameTree{
			{
				Frame: page.Frame{
					ID: "child-a",
				},
				ChildFrames: []page.FrameTree{
					{
						Frame: page.Frame{
							ID: "grandchild-a1",
						},
					},
				},
			},
			{
				Frame: page.Frame{
					ID: "child-b",
				},
			},
		},
	}

	got := collectFrameIDs(tree)

	for _, id := range []page.FrameID{"root", "child-a", "grandchild-a1", "child-b"} {
		if _, ok := got[id]; !ok {
			t.Fatalf("missing frame id %q", id)
		}
	}

	if len(got) != 4 {
		t.Fatalf("unexpected frame id count: %d", len(got))
	}
}
