package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

type recordingDriver struct {
	page        drivers.HTMLPage
	openParams  *drivers.Params
	parseParams *drivers.ParseParams
}

func newRecordingDriver(t *testing.T) (context.Context, *recordingDriver) {
	t.Helper()

	drv := &recordingDriver{page: newTestPage(t, "<html></html>")}
	container := drivers.NewContainer()

	if err := container.Register(drv); err != nil {
		t.Fatal(err)
	}

	container.SetDefault(drv.Name())

	return container.WithContext(t.Context()), drv
}

func (d *recordingDriver) Name() string {
	return "recording"
}

func (d *recordingDriver) Open(_ context.Context, params drivers.Params) (drivers.HTMLPage, error) {
	d.openParams = &params

	return d.page, nil
}

func (d *recordingDriver) Parse(_ context.Context, params drivers.ParseParams) (drivers.HTMLPage, error) {
	d.parseParams = &params

	return d.page, nil
}

func (d *recordingDriver) Close() error {
	return nil
}
