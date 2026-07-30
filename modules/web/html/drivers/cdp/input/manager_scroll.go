package input

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) ScrollTop(ctx context.Context, options drivers.ScrollOptions) (runtime.Boolean, error) {
	m.logger.Trace().
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to the top")

	out, err := m.exec.EvalValue(ctx, templates.ScrollTop(options))
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to the top")

		return runtime.False, err
	}

	m.logger.Trace().Msg("scrolled to the top")

	return runtime.ToBoolean(out), nil
}

func (m *Manager) ScrollBottom(ctx context.Context, options drivers.ScrollOptions) (runtime.Boolean, error) {
	m.logger.Trace().
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to the bottom")

	out, err := m.exec.EvalValue(ctx, templates.ScrollBottom(options))
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to the bottom")

		return runtime.False, err
	}

	m.logger.Trace().Msg("scrolled to the bottom")

	return runtime.ToBoolean(out), nil
}

func (m *Manager) ScrollIntoView(ctx context.Context, id cdpruntime.RemoteObjectID, options drivers.ScrollOptions) error {
	m.logger.Trace().
		Str("object_id", string(id)).
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to an element")

	if err := m.exec.Eval(ctx, templates.ScrollIntoView(id, options)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to an element")

		return err
	}

	m.logger.Trace().Msg("scrolled to an element")

	return nil
}

func (m *Manager) ScrollIntoViewBySelector(
	ctx context.Context,
	id cdpruntime.RemoteObjectID,
	selector drivers.QuerySelector,
	options drivers.ScrollOptions,
) (runtime.Boolean, error) {
	m.logger.Trace().
		Str("selector", selector.String()).
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to an element by selector")

	out, err := m.exec.EvalValue(ctx, templates.ScrollIntoViewBySelector(id, selector, options))
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to an element by selector")

		return runtime.False, err
	}

	m.logger.Trace().Msg("scrolled to an element by selector")

	return runtime.ToBoolean(out), nil
}

func (m *Manager) ScrollByXY(ctx context.Context, options drivers.ScrollOptions) (runtime.Boolean, error) {
	m.logger.Trace().
		Float64("x", float64(options.Left)).
		Float64("y", float64(options.Top)).
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to an element by given coordinates")

	out, err := m.exec.EvalValue(ctx, templates.Scroll(options))
	if err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to an element by coordinates")

		return runtime.False, err
	}

	m.logger.Trace().Msg("scrolled to an element by given coordinates")

	return runtime.ToBoolean(out), nil
}

func (m *Manager) ScrollByDelta(ctx context.Context, options drivers.ScrollOptions) error {
	m.logger.Trace().
		Float64("x", float64(options.Left)).
		Float64("y", float64(options.Top)).
		Str("behavior", options.Behavior.String()).
		Msg("scrolling by given coordinates")

	if err := m.exec.Eval(ctx, templates.ScrollBy(options)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll by coordinates")

		return err
	}

	m.logger.Trace().Msg("scrolled by given coordinates")

	return nil
}
