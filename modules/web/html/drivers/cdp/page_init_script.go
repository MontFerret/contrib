package cdp

import (
	"context"
	"strings"

	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) registerInitScript(ctx context.Context) error {
	if p.initScript == nil || p.initScript.Timing != drivers.InitScriptBeforeDocument {
		return nil
	}

	_, err := p.client.Page.AddScriptToEvaluateOnNewDocument(
		ctx,
		page.NewAddScriptToEvaluateOnNewDocumentArgs(p.initScript.Source),
	)
	if err != nil {
		return runtime.Error(err, "register initScript")
	}

	return nil
}

func (p *HTMLPage) evaluateInitScript(ctx context.Context) error {
	if p.initScript == nil || p.initScript.Timing != drivers.InitScriptAfterNavigation {
		return nil
	}

	reply, err := p.client.Runtime.Evaluate(ctx, cdpruntime.NewEvaluateArgs(p.initScript.Source))
	if err != nil {
		return runtime.Error(err, "evaluate initScript")
	}

	if reply.ExceptionDetails == nil {
		return nil
	}

	message := reply.ExceptionDetails.Text
	if reply.ExceptionDetails.Exception != nil && reply.ExceptionDetails.Exception.Description != nil {
		message = *reply.ExceptionDetails.Exception.Description
	}

	if strings.TrimSpace(message) == "" {
		message = "initScript evaluation failed"
	}

	return runtime.Error(runtime.ErrUnexpected, message)
}
