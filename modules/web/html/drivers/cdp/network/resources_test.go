package network

import (
	"testing"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"
)

func TestNormalizedResourceTypesAreAccepted(t *testing.T) {
	tests := []struct {
		name         string
		resourceType cdpnetwork.ResourceType
	}{
		{"document", cdpnetwork.ResourceTypeDocument},
		{"stylesheet", cdpnetwork.ResourceTypeStylesheet},
		{"image", cdpnetwork.ResourceTypeImage},
		{"media", cdpnetwork.ResourceTypeMedia},
		{"font", cdpnetwork.ResourceTypeFont},
		{"script", cdpnetwork.ResourceTypeScript},
		{"texttrack", cdpnetwork.ResourceTypeTextTrack},
		{"xhr", cdpnetwork.ResourceTypeXHR},
		{"fetch", cdpnetwork.ResourceTypeFetch},
		{"prefetch", cdpnetwork.ResourceTypePrefetch},
		{"eventsource", cdpnetwork.ResourceTypeEventSource},
		{"websocket", cdpnetwork.ResourceTypeWebSocket},
		{"manifest", cdpnetwork.ResourceTypeManifest},
		{"signedexchange", cdpnetwork.ResourceTypeSignedExchange},
		{"ping", cdpnetwork.ResourceTypePing},
		{"cspviolationreport", cdpnetwork.ResourceTypeCSPViolationReport},
		{"preflight", cdpnetwork.ResourceTypePreflight},
		{"other", cdpnetwork.ResourceTypeOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalized := normalizeResourceType(tc.resourceType)
			if normalized == "" {
				t.Fatalf("expected normalized resource type for %s", tc.resourceType)
			}

			if got := toResourceType(normalized); got != tc.resourceType {
				t.Fatalf("expected %q to map back to %s, got %s", normalized, tc.resourceType, got)
			}
		})
	}
}
