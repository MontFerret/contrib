package lib

import (
	"context"
	"io"

	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/logging"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

// Pagination creates an iterator that goes through pages using a CSS selector.
// The iterator starts from the current page i.e. it does not change the page on 1st iteration.
// That allows you to keep scraping logic inside FOR loop.
// @param {HTMLPage} page - Target html page.
// @param {String} selector - CSS selector for a pagination on the page.
func Pagination(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 2)

	if err != nil {
		return runtime.None, err
	}

	page, err := drivers.ToPage(args[0])

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	logger := common.LoggerWithName(logging.From(ctx).With(), "stdlib_html_pagination").
		Str("selector", selector.String()).
		Logger()

	return sdk.NewProxy(&Paging{logger, page, selector}), nil
}

type (
	Paging struct {
		logger   zerolog.Logger
		page     drivers.HTMLPage
		selector drivers.QuerySelector
	}

	PagingIterator struct {
		logger   zerolog.Logger
		page     drivers.HTMLPage
		selector drivers.QuerySelector
		pos      runtime.Int
	}
)

func (p *Paging) Iterate(_ context.Context) (runtime.Iterator, error) {
	return &PagingIterator{p.logger, p.page, p.selector, -1}, nil
}

func (i *PagingIterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	i.pos++

	i.logger.Trace().Int("position", int(i.pos)).Msg("starting to advance iteration")

	if i.pos == 0 {
		i.logger.Trace().Msg("starting point of pagination. nothing to do. exit")
		return runtime.ZeroInt, runtime.ZeroInt, nil
	}

	i.logger.Trace().Msg("checking if an element exists...")
	frame := i.page.GetMainFrame()
	exists, err := frame.ExistsBySelector(ctx, i.selector)

	if err != nil {
		i.logger.Trace().Err(err).Msg("failed to check")

		return runtime.None, runtime.None, err
	}

	if !exists {
		i.logger.Trace().Bool("exists", bool(exists)).Msg("element does not exist. exit")

		return runtime.None, runtime.None, io.EOF
	}

	i.logger.Trace().Bool("exists", bool(exists)).Msg("element exists. clicking...")

	target, err := toRootInteractionTarget(frame)
	if err != nil {
		i.logger.Trace().Err(err).Msg("interaction capability is not supported. exit")

		return runtime.None, runtime.None, err
	}

	err = target.ClickBySelector(ctx, i.selector, 1)

	if err != nil {
		i.logger.Trace().Err(err).Msg("failed to click. exit")

		return runtime.None, runtime.None, err
	}

	i.logger.Trace().Msg("successfully clicked on element. iteration has succeeded")

	// terminate
	return i.pos, i.pos, nil
}
