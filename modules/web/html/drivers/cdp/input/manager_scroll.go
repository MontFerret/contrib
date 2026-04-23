package input

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
)

func (m *Manager) ScrollTop(ctx context.Context, options drivers.ScrollOptions) error {
	m.logger.Trace().
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to the top")

	if err := m.exec.Eval(ctx, templates.ScrollTop(options)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to the top")

		return err
	}

	m.logger.Trace().Msg("scrolled to the top")

	return nil
}

func (m *Manager) ScrollBottom(ctx context.Context, options drivers.ScrollOptions) error {
	m.logger.Trace().
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to the bottom")

	if err := m.exec.Eval(ctx, templates.ScrollBottom(options)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to the bottom")

		return err
	}

	m.logger.Trace().Msg("scrolled to the bottom")

	return nil
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

func (m *Manager) ScrollIntoViewBySelector(ctx context.Context, id cdpruntime.RemoteObjectID, selector drivers.QuerySelector, options drivers.ScrollOptions) error {
	m.logger.Trace().
		Str("selector", selector.String()).
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to an element by selector")

	if err := m.exec.Eval(ctx, templates.ScrollIntoViewBySelector(id, selector, options)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to an element by selector")

		return err
	}

	m.logger.Trace().Msg("scrolled to an element by selector")

	return nil
}

func (m *Manager) ScrollByXY(ctx context.Context, options drivers.ScrollOptions) error {
	m.logger.Trace().
		Float64("x", float64(options.Top)).
		Float64("y", float64(options.Left)).
		Str("behavior", options.Behavior.String()).
		Str("block", options.Block.String()).
		Str("inline", options.Inline.String()).
		Msg("scrolling to an element by given coordinates")

	if err := m.exec.Eval(ctx, templates.Scroll(options)); err != nil {
		m.logger.Trace().Err(err).Msg("failed to scroll to an element by coordinates")

		return err
	}

	m.logger.Trace().Msg("scrolled to an element by given coordinates")

	return nil
}
