package network

import (
	"github.com/mafredri/cdp"
	cdpnetwork "github.com/mafredri/cdp/protocol/network"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

const (
	rootSessionKey              = "root"
	networkSessionDetachedEvent = "network.session_detached"
)

type (
	networkEventFields struct {
		documentURL                 *string
		urlFragment                 *string
		hasPostData                 *bool
		initialPriority             *string
		referrerPolicy              *string
		isLinkPreload               *bool
		isSameSite                  *bool
		hasUserGesture              *bool
		initiatorType               *string
		initiatorURL                *string
		initiatorLineNumber         *float64
		initiatorColumnNumber       *float64
		initiatorRequestID          *string
		charset                     *string
		connectionReused            *bool
		connectionID                *float64
		remoteIPAddress             *string
		remotePort                  *int
		fromEarlyHints              *bool
		responseTime                *float64
		protocol                    *string
		securityState               *string
		cacheStorageCacheName       *string
		serviceWorkerResponseSource *string
		alternateProtocolUsage      *string
		redirectURL                 *string
		redirectStatus              *int
		redirectStatusText          *string
		redirected                  bool
	}

	networkEvent struct {
		networkEventFields
		headers           *drivers.HTTPHeaders
		requestHeaders    *drivers.HTTPHeaders
		client            *cdp.Client
		err               error
		name              string
		sessionKey        string
		requestID         cdpnetwork.RequestID
		loaderID          cdpnetwork.LoaderID
		frameID           string
		url               string
		method            string
		resourceType      string
		statusText        string
		mimeType          string
		errorText         string
		blockedReason     string
		status            int
		failed            bool
		canceled          bool
		fromCache         bool
		fromDiskCache     bool
		fromServiceWorker bool
		fromPrefetchCache bool
		encodedDataLength float64
		timestamp         float64
		wallTime          float64
	}

	networkRequestState struct {
		networkEventFields
		headers           *drivers.HTTPHeaders
		requestHeaders    *drivers.HTTPHeaders
		client            *cdp.Client
		sessionKey        string
		requestID         cdpnetwork.RequestID
		loaderID          cdpnetwork.LoaderID
		frameID           string
		url               string
		method            string
		resourceType      string
		statusText        string
		mimeType          string
		errorText         string
		blockedReason     string
		status            int
		failed            bool
		canceled          bool
		fromCache         bool
		fromDiskCache     bool
		fromServiceWorker bool
		fromPrefetchCache bool
		encodedDataLength float64
		timestamp         float64
		wallTime          float64
	}

	networkEventSubscriber struct {
		ch   chan networkEvent
		done chan struct{}
		id   int64
	}
)

func networkRequestKey(sessionKey string, requestID cdpnetwork.RequestID) string {
	return sessionKey + "\x00" + string(requestID)
}

func networkEventFromState(name string, state networkRequestState) networkEvent {
	return networkEvent{
		networkEventFields: state.networkEventFields,
		name:               name,
		sessionKey:         state.sessionKey,
		requestID:          state.requestID,
		loaderID:           state.loaderID,
		frameID:            state.frameID,
		url:                state.url,
		method:             state.method,
		resourceType:       state.resourceType,
		status:             state.status,
		statusText:         state.statusText,
		mimeType:           state.mimeType,
		headers:            state.headers,
		requestHeaders:     state.requestHeaders,
		failed:             state.failed,
		errorText:          state.errorText,
		canceled:           state.canceled,
		blockedReason:      state.blockedReason,
		fromCache:          state.fromCache,
		fromDiskCache:      state.fromDiskCache,
		fromServiceWorker:  state.fromServiceWorker,
		fromPrefetchCache:  state.fromPrefetchCache,
		encodedDataLength:  state.encodedDataLength,
		timestamp:          state.timestamp,
		wallTime:           state.wallTime,
		client:             state.client,
	}
}
