package network

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	cdpnetwork "github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/security"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type networkEventTestAPI struct {
	cdp.Network
	body *cdpnetwork.GetResponseBodyReply
	err  error
}

func (api *networkEventTestAPI) GetResponseBody(
	context.Context,
	*cdpnetwork.GetResponseBodyArgs,
) (*cdpnetwork.GetResponseBodyReply, error) {
	return api.body, api.err
}

func TestNetworkEventOptionParsing(t *testing.T) {
	ctx := context.Background()

	defaults, err := parseNetworkEventOptions(ctx, drivers.NetworkRequestFinishedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected default option error: %v", err)
	}

	if defaults.captureBody {
		t.Fatal("captureBody should default to false")
	}

	if defaults.bodyLimit != defaultNetworkBodyLimit {
		t.Fatalf("expected bodyLimit %d, got %d", defaultNetworkBodyLimit, defaults.bodyLimit)
	}

	opts := runtime.NewObjectWith(map[string]runtime.Value{
		"captureBody": runtime.True,
		"bodyLimit":   runtime.NewInt(128),
	})

	parsed, err := parseNetworkEventOptions(ctx, drivers.NetworkRequestFinishedEvent, opts)
	if err != nil {
		t.Fatalf("unexpected option error: %v", err)
	}

	if !parsed.captureBody || parsed.bodyLimit != 128 {
		t.Fatalf("unexpected parsed options: %+v", parsed)
	}

	if _, err := parseNetworkEventOptions(
		ctx,
		drivers.NetworkRequestFinishedEvent,
		runtime.NewObjectWith(map[string]runtime.Value{"unknown": runtime.True}),
	); err == nil {
		t.Fatal("expected unknown option to fail")
	}

	if _, err := parseNetworkEventOptions(
		ctx,
		drivers.NetworkRequestStartedEvent,
		runtime.NewObjectWith(map[string]runtime.Value{"captureBody": runtime.True}),
	); err == nil {
		t.Fatal("expected request_started captureBody option to fail")
	}

	idle, err := parseNetworkIdleOptions(ctx, drivers.NetworkIdleEvent, runtime.NewObjectWith(map[string]runtime.Value{
		"quiet":       runtime.NewInt(25),
		"maxInflight": runtime.NewInt(1),
		"types": runtime.NewArrayWith(
			runtime.NewString("XHR"),
			runtime.NewString("fetch"),
			runtime.NewString("ajax"),
			runtime.NewString("Prefetch"),
			runtime.NewString("SignedExchange"),
			runtime.NewString("sxg"),
		),
	}))
	if err != nil {
		t.Fatalf("unexpected idle option error: %v", err)
	}

	if idle.quiet != 25*time.Millisecond || idle.maxInflight != 1 {
		t.Fatalf("unexpected idle timing options: %+v", idle)
	}

	if len(idle.types) != 4 {
		t.Fatalf("expected deduplicated types, got %+v", idle.types)
	}

	if _, exists := idle.types["xhr"]; !exists {
		t.Fatal("expected xhr type")
	}

	if _, exists := idle.types["fetch"]; !exists {
		t.Fatal("expected fetch type")
	}

	if _, exists := idle.types["prefetch"]; !exists {
		t.Fatal("expected prefetch type")
	}

	if _, exists := idle.types["signedexchange"]; !exists {
		t.Fatal("expected signedexchange type")
	}

	expectedTypeList := []string{"xhr", "fetch", "prefetch", "signedexchange"}
	if len(idle.typeList) != len(expectedTypeList) {
		t.Fatalf("expected normalized type list %+v, got %+v", expectedTypeList, idle.typeList)
	}

	for i, expected := range expectedTypeList {
		if idle.typeList[i] != expected {
			t.Fatalf("expected normalized type list %+v, got %+v", expectedTypeList, idle.typeList)
		}
	}
}

func TestNetworkObserverSubscribeRejectsUnknownEventName(t *testing.T) {
	observer := newNetworkObserver(zerolog.Nop(), nil, nil)

	_, err := observer.Subscribe(context.Background(), "network.unknown", nil)
	if err == nil {
		t.Fatal("expected unknown event error")
	}

	if !strings.Contains(err.Error(), "unknown event name: network.unknown") {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEvents := strings.Join(drivers.SupportedNetworkEvents(), ", ")
	if !strings.Contains(err.Error(), "supported events: "+expectedEvents) {
		t.Fatalf("expected supported event list in error, got %v", err)
	}
}

func TestNetworkObserverEmitsRequestStartedMetadata(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkRequestStartedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	frameID := page.FrameID("frame-1")
	hasPostData := true
	isLinkPreload := true
	isSameSite := false
	hasUserGesture := true
	urlFragment := "#details"
	initiatorURL := "https://example.com/app.js"
	initiatorLine := 12.5
	initiatorColumn := 4.25
	initiatorRequestID := cdpnetwork.RequestID("preflight-1")

	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID:      "request-1",
		LoaderID:       "loader-1",
		DocumentURL:    "https://example.com/app",
		FrameID:        &frameID,
		Type:           cdpnetwork.ResourceTypeFetch,
		HasUserGesture: &hasUserGesture,
		Request: cdpnetwork.Request{
			URL:             "https://example.com/api/products",
			URLFragment:     &urlFragment,
			Method:          "POST",
			Headers:         cdpnetwork.Headers(`{"Accept":"application/json"}`),
			HasPostData:     &hasPostData,
			InitialPriority: cdpnetwork.ResourcePriorityHigh,
			ReferrerPolicy:  "strict-origin-when-cross-origin",
			IsLinkPreload:   &isLinkPreload,
			IsSameSite:      &isSameSite,
		},
		Initiator: cdpnetwork.Initiator{
			Type:         "script",
			URL:          &initiatorURL,
			LineNumber:   &initiatorLine,
			ColumnNumber: &initiatorColumn,
			RequestID:    &initiatorRequestID,
		},
		Timestamp: 1.25,
		WallTime:  10,
	})

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkRequestStartedEvent)
	assertStringField(t, payload, "documentURL", "https://example.com/app")
	assertStringField(t, payload, "urlFragment", "#details")
	assertBoolField(t, payload, "hasPostData", true)
	assertStringField(t, payload, "initialPriority", "High")
	assertStringField(t, payload, "referrerPolicy", "strict-origin-when-cross-origin")
	assertBoolField(t, payload, "isLinkPreload", true)
	assertBoolField(t, payload, "isSameSite", false)
	assertBoolField(t, payload, "hasUserGesture", true)
	assertStringField(t, payload, "initiatorType", "script")
	assertStringField(t, payload, "initiatorURL", "https://example.com/app.js")
	assertFloatField(t, payload, "initiatorLineNumber", 12.5)
	assertFloatField(t, payload, "initiatorColumnNumber", 4.25)
	assertStringField(t, payload, "initiatorRequestId", "preflight-1")
	assertBoolField(t, payload, "redirected", false)
	assertNoneField(t, payload, "redirectURL")
}

func TestNetworkObserverEmitsRedirectMetadata(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkRequestStartedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID:   "request-2",
		DocumentURL: "https://example.com/next",
		Type:        cdpnetwork.ResourceTypeDocument,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/next",
			Method: "GET",
		},
		RedirectResponse: &cdpnetwork.Response{
			URL:        "https://example.com/old",
			Status:     302,
			StatusText: "Found",
		},
	})

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkRequestStartedEvent)
	assertBoolField(t, payload, "redirected", true)
	assertStringField(t, payload, "redirectURL", "https://example.com/old")
	assertIntField(t, payload, "redirectStatus", 302)
	assertStringField(t, payload, "redirectStatusText", "Found")
}

func TestNetworkObserverEmitsRequestResponseAndFinishedPayloads(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkRequestFinishedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	frameID := page.FrameID("frame-1")
	hasPostData := true
	isSameSite := true
	remoteIP := "203.0.113.10"
	remotePort := 443
	fromEarlyHints := true
	protocol := "h2"
	cacheName := "api-cache"
	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID:   "request-1",
		LoaderID:    "loader-1",
		DocumentURL: "https://example.com/app",
		FrameID:     &frameID,
		Type:        cdpnetwork.ResourceTypeFetch,
		Request: cdpnetwork.Request{
			URL:             "https://example.com/api/products",
			Method:          "POST",
			Headers:         cdpnetwork.Headers(`{"Accept":"application/json"}`),
			HasPostData:     &hasPostData,
			InitialPriority: cdpnetwork.ResourcePriorityVeryHigh,
			ReferrerPolicy:  "origin",
			IsSameSite:      &isSameSite,
		},
		Timestamp: 1.25,
		WallTime:  10,
	})
	observer.handleResponseReceived(rootSessionKey, client, &cdpnetwork.ResponseReceivedReply{
		RequestID: "request-1",
		LoaderID:  "loader-1",
		FrameID:   &frameID,
		Type:      cdpnetwork.ResourceTypeFetch,
		Response: cdpnetwork.Response{
			URL:                         "https://example.com/api/products",
			Status:                      200,
			StatusText:                  "OK",
			MimeType:                    "application/json",
			Charset:                     "utf-8",
			Headers:                     cdpnetwork.Headers(`{"Content-Type":"application/json"}`),
			RequestHeaders:              cdpnetwork.Headers(`{"Accept":"application/json"}`),
			ConnectionReused:            true,
			ConnectionID:                7,
			RemoteIPAddress:             &remoteIP,
			RemotePort:                  &remotePort,
			FromEarlyHints:              &fromEarlyHints,
			EncodedDataLength:           42,
			ResponseTime:                cdpnetwork.TimeSinceEpoch(1234.5),
			Protocol:                    &protocol,
			SecurityState:               security.StateSecure,
			CacheStorageCacheName:       &cacheName,
			ServiceWorkerResponseSource: cdpnetwork.ServiceWorkerResponseSourceNetwork,
			AlternateProtocolUsage:      cdpnetwork.AlternateProtocolUsageAlternativeJobWonRace,
		},
		Timestamp: 2.5,
	})
	observer.handleRequestFinished(rootSessionKey, client, &cdpnetwork.LoadingFinishedReply{
		RequestID:         "request-1",
		Timestamp:         3.75,
		EncodedDataLength: 128,
	})

	payload := readNetworkObject(t, messages)

	assertStringField(t, payload, "event", drivers.NetworkRequestFinishedEvent)
	assertStringField(t, payload, "requestId", "request-1")
	assertStringField(t, payload, "loaderId", "loader-1")
	assertStringField(t, payload, "frameId", "frame-1")
	assertStringField(t, payload, "url", "https://example.com/api/products")
	assertStringField(t, payload, "method", "POST")
	assertStringField(t, payload, "type", "fetch")
	assertIntField(t, payload, "status", 200)
	assertStringField(t, payload, "statusText", "OK")
	assertStringField(t, payload, "mimeType", "application/json")
	assertBoolField(t, payload, "failed", false)
	assertBoolField(t, payload, "fromCache", false)
	assertFloatField(t, payload, "encodedDataLength", 128)
	assertStringField(t, payload, "documentURL", "https://example.com/app")
	assertBoolField(t, payload, "hasPostData", true)
	assertStringField(t, payload, "initialPriority", "VeryHigh")
	assertStringField(t, payload, "referrerPolicy", "origin")
	assertBoolField(t, payload, "isSameSite", true)
	assertStringField(t, payload, "charset", "utf-8")
	assertBoolField(t, payload, "connectionReused", true)
	assertFloatField(t, payload, "connectionId", 7)
	assertStringField(t, payload, "remoteIPAddress", "203.0.113.10")
	assertIntField(t, payload, "remotePort", 443)
	assertBoolField(t, payload, "fromEarlyHints", true)
	assertFloatField(t, payload, "responseTime", 1234.5)
	assertStringField(t, payload, "protocol", "h2")
	assertStringField(t, payload, "securityState", "secure")
	assertStringField(t, payload, "cacheStorageCacheName", "api-cache")
	assertStringField(t, payload, "serviceWorkerResponseSource", "network")
	assertStringField(t, payload, "alternateProtocolUsage", "alternativeJobWonRace")
}

func TestNetworkObserverEmitsResponseReceivedPayload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkResponseReceivedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	fromEarlyHints := false
	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID: "request-1",
		Type:      cdpnetwork.ResourceTypeXHR,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/api/products",
			Method: "GET",
		},
	})
	observer.handleResponseReceived(rootSessionKey, client, &cdpnetwork.ResponseReceivedReply{
		RequestID: "request-1",
		Type:      cdpnetwork.ResourceTypeXHR,
		Response: cdpnetwork.Response{
			URL:            "https://example.com/api/products",
			Status:         204,
			StatusText:     "No Content",
			MimeType:       "application/json",
			Charset:        "utf-8",
			FromEarlyHints: &fromEarlyHints,
			SecurityState:  security.StateNeutral,
		},
	})

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkResponseReceivedEvent)
	assertStringField(t, payload, "type", "xhr")
	assertIntField(t, payload, "status", 204)
	assertStringField(t, payload, "statusText", "No Content")
	assertStringField(t, payload, "charset", "utf-8")
	assertBoolField(t, payload, "connectionReused", false)
	assertBoolField(t, payload, "fromEarlyHints", false)
	assertStringField(t, payload, "securityState", "neutral")
}

func TestNetworkObserverEmitsSignedExchangeResourceType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkRequestStartedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID: "request-1",
		Type:      cdpnetwork.ResourceTypeSignedExchange,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/page.sxg",
			Method: "GET",
		},
	})

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkRequestStartedEvent)
	assertStringField(t, payload, "type", "signedexchange")
}

func TestNetworkObserverEmitsRequestFailedPayload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkRequestFailedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID:   "request-1",
		DocumentURL: "https://example.com/app",
		Type:        cdpnetwork.ResourceTypeFetch,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/api/products",
			Method: "GET",
		},
	})

	canceled := true
	observer.handleRequestFailed(rootSessionKey, client, &cdpnetwork.LoadingFailedReply{
		RequestID:     "request-1",
		Type:          cdpnetwork.ResourceTypeFetch,
		ErrorText:     "net::ERR_ABORTED",
		Canceled:      &canceled,
		BlockedReason: cdpnetwork.BlockedReasonInspector,
	})

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkRequestFailedEvent)
	assertStringField(t, payload, "type", "fetch")
	assertBoolField(t, payload, "failed", true)
	assertBoolField(t, payload, "canceled", true)
	assertStringField(t, payload, "errorText", "net::ERR_ABORTED")
	assertStringField(t, payload, "blockedReason", "inspector")
	assertStringField(t, payload, "documentURL", "https://example.com/app")
	assertNoneField(t, payload, "connectionReused")
	assertNoneField(t, payload, "remoteIPAddress")
}

func TestNetworkObserverPropagatesFromCacheState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkRequestFinishedEvent, nil)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID: "request-1",
		Type:      cdpnetwork.ResourceTypeImage,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/logo.png",
			Method: "GET",
		},
	})
	observer.handleRequestServedFromCache(rootSessionKey, &cdpnetwork.RequestServedFromCacheReply{
		RequestID: "request-1",
	})
	observer.handleRequestFinished(rootSessionKey, client, &cdpnetwork.LoadingFinishedReply{
		RequestID: "request-1",
	})

	payload := readNetworkObject(t, messages)
	assertBoolField(t, payload, "fromCache", true)
	assertNoneField(t, payload, "fromEarlyHints")
}

func TestNetworkObserverSlowSubscriberDoesNotBlockDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	observer := newStartedTestNetworkObserver(ctx, nil)

	blocked := observer.subscribe()
	defer observer.unsubscribe(blocked.id)

	receiver := observer.subscribe()
	defer observer.unsubscribe(receiver.id)

	for i := 0; i < cap(blocked.ch); i++ {
		blocked.ch <- networkEvent{name: drivers.NetworkRequestStartedEvent}
	}

	event := networkEvent{
		name:      drivers.NetworkRequestStartedEvent,
		requestID: "request-1",
	}

	done := make(chan struct{})
	go func() {
		observer.emit(event)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("emit blocked on a full subscriber channel")
	}

	select {
	case received := <-receiver.ch:
		if received.requestID != event.requestID {
			t.Fatalf("expected request ID %q, got %q", event.requestID, received.requestID)
		}
	case <-time.After(time.Second):
		t.Fatal("expected non-blocked subscriber to receive event")
	}
}

func TestNetworkEventPayloadCapturesAndTruncatesBody(t *testing.T) {
	client := &cdp.Client{Network: &networkEventTestAPI{
		body: &cdpnetwork.GetResponseBodyReply{Body: "abcdef"},
	}}

	payload := buildNetworkEventPayload(
		context.Background(),
		zerolog.Nop(),
		networkEvent{
			name:      drivers.NetworkRequestFinishedEvent,
			requestID: "request-1",
			client:    client,
		},
		networkEventOptions{
			captureBody: true,
			bodyLimit:   3,
		},
	).(*runtime.Object)

	bodyValue, err := payload.Get(context.Background(), runtime.NewString("body"))
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	body, ok := bodyValue.(runtime.Binary)
	if !ok {
		t.Fatalf("expected binary body, got %T", bodyValue)
	}

	if string(body) != "abc" {
		t.Fatalf("expected truncated body abc, got %q", string(body))
	}

	assertBoolField(t, payload, "bodyTruncated", true)
}

func TestNetworkEventPayloadUsesNoneForUnavailableOptionalFields(t *testing.T) {
	payload := buildNetworkEventPayload(
		context.Background(),
		zerolog.Nop(),
		networkEvent{name: drivers.NetworkRequestStartedEvent},
		networkEventOptions{},
	).(*runtime.Object)

	for _, field := range []string{
		"documentURL",
		"urlFragment",
		"hasPostData",
		"initialPriority",
		"referrerPolicy",
		"isLinkPreload",
		"isSameSite",
		"hasUserGesture",
		"initiatorType",
		"initiatorURL",
		"initiatorLineNumber",
		"initiatorColumnNumber",
		"initiatorRequestId",
		"charset",
		"connectionReused",
		"connectionId",
		"remoteIPAddress",
		"remotePort",
		"fromEarlyHints",
		"responseTime",
		"protocol",
		"securityState",
		"cacheStorageCacheName",
		"serviceWorkerResponseSource",
		"alternateProtocolUsage",
		"redirectURL",
		"redirectStatus",
		"redirectStatusText",
	} {
		assertNoneField(t, payload, field)
	}

	assertBoolField(t, payload, "redirected", false)
}

func TestNetworkIdleWaitsForScopedRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID: "fetch-1",
		Type:      cdpnetwork.ResourceTypeFetch,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/api/products",
			Method: "GET",
		},
	})
	observer.handleRequestStarted(rootSessionKey, client, &cdpnetwork.RequestWillBeSentReply{
		RequestID: "xhr-1",
		Type:      cdpnetwork.ResourceTypeXHR,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/ignored",
			Method: "GET",
		},
	})

	stream, err := observer.Subscribe(ctx, drivers.NetworkIdleEvent, runtime.NewObjectWith(map[string]runtime.Value{
		"quiet":       runtime.NewInt(10),
		"maxInflight": runtime.NewInt(0),
		"types":       runtime.NewArrayWith(runtime.NewString("fetch")),
	}))
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	observer.handleRequestFinished(rootSessionKey, client, &cdpnetwork.LoadingFinishedReply{
		RequestID: "xhr-1",
	})

	assertNoNetworkMessage(t, messages, 20*time.Millisecond)

	observer.handleRequestFinished(rootSessionKey, client, &cdpnetwork.LoadingFinishedReply{
		RequestID: "fetch-1",
	})

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkIdleEvent)
	assertIntField(t, payload, "inflight", 0)
	assertIntField(t, payload, "maxInflight", 0)
}

func TestNetworkIdleCleansUpDetachedSessionRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	observer.handleRequestStarted("session-1", client, &cdpnetwork.RequestWillBeSentReply{
		RequestID: "fetch-1",
		Type:      cdpnetwork.ResourceTypeFetch,
		Request: cdpnetwork.Request{
			URL:    "https://example.com/api/products",
			Method: "GET",
		},
	})

	stream, err := observer.Subscribe(ctx, drivers.NetworkIdleEvent, runtime.NewObjectWith(map[string]runtime.Value{
		"quiet": runtime.NewInt(10),
		"types": runtime.NewArrayWith(runtime.NewString("fetch")),
	}))
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)

	observer.handleSessionDetached("session-1")

	payload := readNetworkObject(t, messages)
	assertStringField(t, payload, "event", drivers.NetworkIdleEvent)
	assertIntField(t, payload, "inflight", 0)
}

func TestNetworkIdleReadExitsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &cdp.Client{Network: &networkEventTestAPI{}}
	observer := newStartedTestNetworkObserver(ctx, client)

	stream, err := observer.Subscribe(ctx, drivers.NetworkIdleEvent, runtime.NewObject())
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer stream.Close()

	messages := stream.Read(ctx)
	waitForNetworkSubscribers(t, observer, 1)
	cancel()

	select {
	case _, ok := <-messages:
		if ok {
			t.Fatal("expected idle stream to close after context cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("idle stream did not close after context cancellation")
	}
}

func newStartedTestNetworkObserver(ctx context.Context, client *cdp.Client) *networkObserver {
	observer := newNetworkObserver(zerolog.Nop(), client, nil)
	observer.ctx, observer.cancel = context.WithCancel(ctx)

	return observer
}

func waitForNetworkSubscribers(t *testing.T, observer *networkObserver, expected int) {
	t.Helper()

	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		observer.mu.Lock()
		actual := len(observer.subscribers)
		observer.mu.Unlock()

		if actual >= expected {
			return
		}

		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %d subscribers, got %d", expected, actual)
		case <-ticker.C:
		}
	}
}

func readNetworkObject(t *testing.T, messages <-chan runtime.Message) *runtime.Object {
	t.Helper()

	select {
	case msg := <-messages:
		if err := msg.Err(); err != nil {
			t.Fatalf("unexpected stream error: %v", err)
		}

		obj, ok := msg.Value().(*runtime.Object)
		if !ok {
			t.Fatalf("expected object payload, got %T", msg.Value())
		}

		return obj
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for network message")
	}

	return nil
}

func assertNoNetworkMessage(t *testing.T, messages <-chan runtime.Message, timeout time.Duration) {
	t.Helper()

	select {
	case msg := <-messages:
		t.Fatalf("unexpected network message: %+v", msg.Value())
	case <-time.After(timeout):
	}
}

func assertStringField(t *testing.T, obj *runtime.Object, field string, expected string) {
	t.Helper()

	value := objectField(t, obj, field)
	actual, ok := value.(runtime.String)
	if !ok {
		t.Fatalf("expected %s to be string, got %T", field, value)
	}

	if actual.String() != expected {
		t.Fatalf("expected %s %q, got %q", field, expected, actual)
	}
}

func assertIntField(t *testing.T, obj *runtime.Object, field string, expected int) {
	t.Helper()

	value := objectField(t, obj, field)
	actual, ok := value.(runtime.Int)
	if !ok {
		t.Fatalf("expected %s to be int, got %T", field, value)
	}

	if int(actual) != expected {
		t.Fatalf("expected %s %d, got %d", field, expected, actual)
	}
}

func assertFloatField(t *testing.T, obj *runtime.Object, field string, expected float64) {
	t.Helper()

	value := objectField(t, obj, field)
	actual, ok := value.(runtime.Float)
	if !ok {
		t.Fatalf("expected %s to be float, got %T", field, value)
	}

	if float64(actual) != expected {
		t.Fatalf("expected %s %f, got %f", field, expected, actual)
	}
}

func assertBoolField(t *testing.T, obj *runtime.Object, field string, expected bool) {
	t.Helper()

	value := objectField(t, obj, field)
	actual, ok := value.(runtime.Boolean)
	if !ok {
		t.Fatalf("expected %s to be bool, got %T", field, value)
	}

	if bool(actual) != expected {
		t.Fatalf("expected %s %v, got %v", field, expected, actual)
	}
}

func assertNoneField(t *testing.T, obj *runtime.Object, field string) {
	t.Helper()

	value := objectField(t, obj, field)
	if value != runtime.None {
		t.Fatalf("expected %s to be none, got %T (%v)", field, value, value)
	}
}

func objectField(t *testing.T, obj *runtime.Object, field string) runtime.Value {
	t.Helper()

	value, err := obj.Get(context.Background(), runtime.NewString(field))
	if err != nil {
		t.Fatalf("failed to get %s: %v", field, err)
	}

	return value
}
