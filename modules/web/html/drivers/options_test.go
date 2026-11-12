package drivers_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
)

func TestSharedOptionsSnapshot(t *testing.T) {
	const allowed = "https://listed.example"
	origins := []string{allowed}
	common := []drivers.Option{
		drivers.WithAllowedOrigins(),
		drivers.WithAllowedOrigins(origins...),
		drivers.WithUserAgent("crawler"),
	}
	static := memory.WithOptions(common...)
	browser := cdp.WithOptions(common...)

	origins[0] = "https://other.example"
	common[1] = drivers.WithAllowedOrigins()

	first := memory.NewOptions([]memory.Option{static})
	second := cdp.NewOptions([]cdp.Option{browser})
	if !assert.Equal(t, []string{allowed}, first.AllowedOrigins) {
		t.FailNow()
	}

	first.AllowedOrigins[0] = "https://changed.example"
	assert.Equal(t, []string{allowed}, second.AllowedOrigins)
	assert.Equal(t, "crawler", first.UserAgent)
	assert.Equal(t, "crawler", second.UserAgent)
	assert.Equal(t, []string{allowed}, memory.NewOptions([]memory.Option{static}).AllowedOrigins)
}

func TestSharedOptionsPrecedence(t *testing.T) {
	const allowed = "https://listed.example"
	denyAll := drivers.WithAllowedOrigins()

	t.Run("memory", func(t *testing.T) {
		bundle := memory.WithOptions(denyAll)
		individual := memory.WithAllowedOrigins(allowed)

		individualLast := memory.NewOptions([]memory.Option{bundle, individual})
		bundleLast := memory.NewOptions([]memory.Option{individual, bundle})
		assert.Equal(t, []string{allowed}, individualLast.AllowedOrigins)
		assert.Equal(t, []string{}, bundleLast.AllowedOrigins)

		omitted := memory.NewOptions([]memory.Option{
			memory.WithHeader("X-Test", []string{"keep"}),
			memory.WithHeaders(nil),
			memory.WithOptions(),
		})
		assert.Nil(t, omitted.AllowedOrigins)

		if assert.NotNil(t, omitted.Headers) {
			assert.Equal(t, "keep", omitted.Headers.Data.Get("X-Test"))
		}
	})

	t.Run("cdp", func(t *testing.T) {
		bundle := cdp.WithOptions(denyAll)
		individual := cdp.WithAllowedOrigins(allowed)

		individualLast := cdp.NewOptions([]cdp.Option{bundle, individual})
		bundleLast := cdp.NewOptions([]cdp.Option{individual, bundle})
		assert.Equal(t, []string{allowed}, individualLast.AllowedOrigins)
		assert.Equal(t, []string{}, bundleLast.AllowedOrigins)
		assert.Nil(t, cdp.NewOptions([]cdp.Option{cdp.WithOptions()}).AllowedOrigins)
	})
}

func TestAllowedOriginsIsHostOnly(t *testing.T) {
	var opts drivers.Options
	drivers.WithAllowedOrigins("https://host.example")(&opts)

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}

	assert.NotContains(t, fields, "AllowedOrigins")
	assert.NotContains(t, fields, "allowedOrigins")

	if err := json.Unmarshal([]byte(`{"allowedOrigins":[]}`), &opts); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, []string{"https://host.example"}, opts.AllowedOrigins)
}
