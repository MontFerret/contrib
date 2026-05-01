package dom

import (
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

type Manager struct {
	logger     zerolog.Logger
	rootClient *cdp.Client
	mouse      *input.Mouse
	keyboard   *input.Keyboard
	mainFrame  *AtomicFrameID
	frames     *AtomicFrameCollection
	owners     *AtomicFrameClientCollection
	mu         sync.RWMutex
}

func New(
	logger zerolog.Logger,
	client *cdp.Client,
	mouse *input.Mouse,
	keyboard *input.Keyboard,
) (manager *Manager, err error) {

	manager = new(Manager)
	manager.logger = logutil.WithComponent(logger.With(), "dom_manager").Logger()
	manager.rootClient = client
	manager.mouse = mouse
	manager.keyboard = keyboard
	manager.mainFrame = NewAtomicFrameID()
	manager.frames = NewAtomicFrameCollection()
	manager.owners = NewAtomicFrameClientCollection()

	return manager, nil
}

func (m *Manager) GetMainFrame() *HTMLDocument {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mainFrameID := m.mainFrame.Get()

	if mainFrameID == "" {
		return nil
	}

	mainFrame, exists := m.frames.Get(mainFrameID)

	if exists {
		return mainFrame.node
	}

	return nil
}

func (m *Manager) RecordFrameClient(frameID page.FrameID, client *cdp.Client) {
	if client == nil || frameID == "" {
		return
	}

	m.owners.Set(frameID, client)
}
