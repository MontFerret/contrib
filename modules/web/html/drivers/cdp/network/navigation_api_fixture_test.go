package network

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
)

type navigationTestAPI struct {
	cdp.Page
	frame         *nilFrameNavigationStream
	within        *nilWithinDocumentNavigationStream
	frameErr      error
	withinErr     error
	history       *page.GetNavigationHistoryReply
	navigate      func() (*page.NavigateReply, error)
	change        func() error
	subscriptions atomic.Int32
	commands      atomic.Int32
}

func newNavigationTestAPI() *navigationTestAPI {
	return &navigationTestAPI{
		frame:  &nilFrameNavigationStream{nilNavigationStream: newNilNavigationStream(false)},
		within: &nilWithinDocumentNavigationStream{newNilNavigationStream(false)},
	}
}

func (api *navigationTestAPI) FrameNavigated(context.Context) (page.FrameNavigatedClient, error) {
	api.subscriptions.Add(1)
	if api.frameErr != nil {
		return nil, api.frameErr
	}

	return api.frame, nil
}

func (api *navigationTestAPI) NavigatedWithinDocument(context.Context) (page.NavigatedWithinDocumentClient, error) {
	if api.withinErr != nil {
		return nil, api.withinErr
	}

	return api.within, nil
}

func (api *navigationTestAPI) Navigate(context.Context, *page.NavigateArgs) (*page.NavigateReply, error) {
	api.commands.Add(1)
	if api.navigate == nil {
		return nil, errors.New("unexpected navigation command")
	}

	return api.navigate()
}

func (api *navigationTestAPI) GetNavigationHistory(context.Context) (*page.GetNavigationHistoryReply, error) {
	return api.history, nil
}

func (api *navigationTestAPI) NavigateToHistoryEntry(context.Context, *page.NavigateToHistoryEntryArgs) error {
	api.commands.Add(1)
	if api.change == nil {
		return errors.New("unexpected history navigation")
	}

	return api.change()
}
