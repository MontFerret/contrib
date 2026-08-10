package cdp

import (
	"context"
	"errors"
	"testing"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type initScriptPageAPI struct {
	cdp.Page
	registered []string
	closeCalls int
}

func (api *initScriptPageAPI) AddScriptToEvaluateOnNewDocument(
	_ context.Context,
	args *page.AddScriptToEvaluateOnNewDocumentArgs,
) (*page.AddScriptToEvaluateOnNewDocumentReply, error) {
	api.registered = append(api.registered, args.Source)

	return &page.AddScriptToEvaluateOnNewDocumentReply{}, nil
}

func (api *initScriptPageAPI) Close(context.Context) error {
	api.closeCalls++

	return nil
}

type initScriptRuntimeAPI struct {
	cdp.Runtime
	reply     *cdpruntime.EvaluateReply
	evaluated []string
}

func (api *initScriptRuntimeAPI) Evaluate(
	_ context.Context,
	args *cdpruntime.EvaluateArgs,
) (*cdpruntime.EvaluateReply, error) {
	api.evaluated = append(api.evaluated, args.Expression)
	if api.reply != nil {
		return api.reply, nil
	}

	return &cdpruntime.EvaluateReply{}, nil
}

func TestInitScriptLifecycleCalls(t *testing.T) {
	ctx := context.Background()

	t.Run("before document registration", func(t *testing.T) {
		pageAPI := new(initScriptPageAPI)
		p := NewHTMLPage(zerolog.Nop(), &cdp.Client{Page: pageAPI}, nil, nil, nil)
		p.initScript = &drivers.InitScript{Source: "window.ready = true", Timing: drivers.InitScriptBeforeDocument}

		if err := p.registerInitScript(ctx); err != nil {
			t.Fatalf("register: %v", err)
		}
		if len(pageAPI.registered) != 1 || pageAPI.registered[0] != p.initScript.Source {
			t.Fatalf("unexpected registrations: %#v", pageAPI.registered)
		}
	})

	t.Run("after navigation evaluation", func(t *testing.T) {
		runtimeAPI := new(initScriptRuntimeAPI)
		p := NewHTMLPage(zerolog.Nop(), &cdp.Client{Runtime: runtimeAPI}, nil, nil, nil)
		p.initScript = &drivers.InitScript{Source: "window.ready = true", Timing: drivers.InitScriptAfterNavigation}

		if err := p.evaluateInitScript(ctx); err != nil {
			t.Fatalf("evaluate: %v", err)
		}
		if len(runtimeAPI.evaluated) != 1 || runtimeAPI.evaluated[0] != p.initScript.Source {
			t.Fatalf("unexpected evaluations: %#v", runtimeAPI.evaluated)
		}
	})

	t.Run("evaluation exception", func(t *testing.T) {
		description := "ReferenceError: missing is not defined"
		runtimeAPI := &initScriptRuntimeAPI{reply: &cdpruntime.EvaluateReply{
			ExceptionDetails: &cdpruntime.ExceptionDetails{
				Text: "Uncaught",
				Exception: &cdpruntime.RemoteObject{
					Description: &description,
				},
			},
		}}
		p := NewHTMLPage(zerolog.Nop(), &cdp.Client{Runtime: runtimeAPI}, nil, nil, nil)
		p.initScript = &drivers.InitScript{Source: "missing", Timing: drivers.InitScriptAfterNavigation}

		err := p.evaluateInitScript(ctx)
		if !errors.Is(err, runtime.ErrUnexpected) {
			t.Fatalf("expected runtime exception, got %v", err)
		}
	})
}

func TestLoadHTMLPageValidatesInitScriptBeforeOpening(t *testing.T) {
	_, err := LoadHTMLPage(context.Background(), nil, drivers.Params{
		InitScript: &drivers.InitScript{Source: " ", Timing: drivers.InitScriptAfterNavigation},
	})
	if !errors.Is(err, runtime.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestHTMLPageCloseIsIdempotent(t *testing.T) {
	pageAPI := new(initScriptPageAPI)
	p := NewHTMLPage(zerolog.Nop(), &cdp.Client{Page: pageAPI}, nil, nil, nil)

	if err := p.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
	if pageAPI.closeCalls != 1 {
		t.Fatalf("close calls = %d, want 1", pageAPI.closeCalls)
	}
}
