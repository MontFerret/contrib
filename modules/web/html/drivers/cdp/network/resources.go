package network

import (
	"strings"

	"github.com/mafredri/cdp/protocol/network"
)

var (
	resourceTypeMapping = map[string]network.ResourceType{
		"document":           network.ResourceTypeDocument,
		"stylesheet":         network.ResourceTypeStylesheet,
		"css":                network.ResourceTypeStylesheet,
		"image":              network.ResourceTypeImage,
		"media":              network.ResourceTypeMedia,
		"font":               network.ResourceTypeFont,
		"script":             network.ResourceTypeScript,
		"js":                 network.ResourceTypeScript,
		"texttrack":          network.ResourceTypeTextTrack,
		"xhr":                network.ResourceTypeXHR,
		"ajax":               network.ResourceTypeXHR,
		"fetch":              network.ResourceTypeFetch,
		"eventsource":        network.ResourceTypeEventSource,
		"websocket":          network.ResourceTypeWebSocket,
		"manifest":           network.ResourceTypeManifest,
		"sxg":                network.ResourceTypeSignedExchange,
		"ping":               network.ResourceTypePing,
		"cspviolationreport": network.ResourceTypeCSPViolationReport,
		"preflight":          network.ResourceTypePreflight,
		"other":              network.ResourceTypeOther,
	}
)

func toResourceType(alias string) network.ResourceType {
	rt, found := resourceTypeMapping[strings.ToLower(alias)]

	if found {
		return rt
	}

	return network.ResourceTypeNotSet
}

func normalizeResourceType(rt network.ResourceType) string {
	switch rt {
	case network.ResourceTypeDocument:
		return "document"
	case network.ResourceTypeStylesheet:
		return "stylesheet"
	case network.ResourceTypeImage:
		return "image"
	case network.ResourceTypeMedia:
		return "media"
	case network.ResourceTypeFont:
		return "font"
	case network.ResourceTypeScript:
		return "script"
	case network.ResourceTypeTextTrack:
		return "texttrack"
	case network.ResourceTypeXHR:
		return "xhr"
	case network.ResourceTypeFetch:
		return "fetch"
	case network.ResourceTypePrefetch:
		return "prefetch"
	case network.ResourceTypeEventSource:
		return "eventsource"
	case network.ResourceTypeWebSocket:
		return "websocket"
	case network.ResourceTypeManifest:
		return "manifest"
	case network.ResourceTypeSignedExchange:
		return "sxg"
	case network.ResourceTypePing:
		return "ping"
	case network.ResourceTypeCSPViolationReport:
		return "cspviolationreport"
	case network.ResourceTypePreflight:
		return "preflight"
	case network.ResourceTypeOther:
		return "other"
	default:
		return strings.ToLower(rt.String())
	}
}

func normalizeResourceTypeAlias(alias string) string {
	rt := toResourceType(alias)
	if rt == network.ResourceTypeNotSet {
		return strings.ToLower(alias)
	}

	return normalizeResourceType(rt)
}
