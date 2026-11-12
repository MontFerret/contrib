package session

import (
	"context"
	"errors"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
)

// ClosePageTargets closes the page and its popup descendants. An optional
// replacement browser client permits cleanup after the page socket is lost.
func (m *Manager) ClosePageTargets(ctx context.Context, browser *cdp.Client) error {
	if m.targets == nil {
		if root := m.Root(); root != nil {
			return root.CDP.Page.Close(ctx)
		}

		return nil
	}

	if browser == nil {
		browser = m.browserClient
	}

	_, err := browser.Target.CloseTarget(ctx, target.NewCloseTargetArgs(m.rootTargetID))
	if err != nil {
		// A page can be closed externally while its popups remain alive.
		reply, checkErr := browser.Target.GetTargets(ctx, target.NewGetTargetsArgs())
		if checkErr != nil {
			return errors.Join(err, checkErr)
		}

		for _, info := range reply.TargetInfos {
			if info.TargetID == m.rootTargetID {
				return errors.Join(err, m.closePopups(ctx, browser))
			}
		}
	}

	return m.closePopups(ctx, browser)
}

func (m *Manager) closePopups(ctx context.Context, browser *cdp.Client) error {
	failed := make(map[target.ID]error)
	for {
		reply, err := browser.Target.GetTargets(ctx, target.NewGetTargetsArgs())
		if err != nil {
			errs := []error{err}
			for _, cause := range failed {
				errs = append(errs, cause)
			}

			return errors.Join(errs...)
		}

		owned := m.targets.livePopups(reply.TargetInfos)
		var stillOpen []error
		for _, id := range owned {
			// The target may have closed itself before our command arrived.
			// Only report a failed close if the next enumeration still sees it.
			if cause := failed[id]; cause != nil {
				stillOpen = append(stillOpen, cause)

				continue
			}

			if _, err := browser.Target.CloseTarget(ctx, target.NewCloseTargetArgs(id)); err != nil {
				failed[id] = err
			}
		}

		if len(stillOpen) == len(owned) {
			return errors.Join(stillOpen...)
		}
	}
}
