package dom

import (
	"errors"

	"github.com/mafredri/cdp/protocol/page"
)

func (m *Manager) Close() error {
	errs := make([]error, 0, m.frames.Length()+1)

	m.frames.ForEach(func(f Frame, _ page.FrameID) bool {
		// if initialized
		if f.node != nil {
			if err := f.node.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		return true
	})

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
