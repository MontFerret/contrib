package input

import (
	"time"

	"github.com/mafredri/cdp"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

type (
	TypeParams struct {
		Text  string
		Clear bool
		Delay time.Duration
	}

	Manager struct {
		logger   zerolog.Logger
		client   *cdp.Client
		exec     *eval.Runtime
		keyboard *Keyboard
		mouse    *Mouse
	}
)

func New(
	logger zerolog.Logger,
	client *cdp.Client,
	exec *eval.Runtime,
	keyboard *Keyboard,
	mouse *Mouse,
) *Manager {
	logger = logutil.WithComponent(logger.With(), "input_manager").Logger()

	return &Manager{
		logger,
		client,
		exec,
		keyboard,
		mouse,
	}
}

func (m *Manager) Keyboard() *Keyboard {
	return m.keyboard
}

func (m *Manager) Mouse() *Mouse {
	return m.mouse
}
