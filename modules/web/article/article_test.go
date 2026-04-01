package article

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MontFerret/contrib/modules/web/article/core"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "web/article" {
		t.Fatalf("expected module name %q, got %q", "web/article", mod.Name())
	}
}

func TestRegisterInstallsSessionExtractor(t *testing.T) {
	mod := New()

	seenExtractor := false
	engine, err := ferret.New(
		ferret.WithModules(mod),
		ferret.WithAfterRunHook(func(ctx context.Context, err error) error {
			if err == nil && core.ExtractorFromContext(ctx) != nil {
				seenExtractor = true
			}

			return nil
		}),
	)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	defer func() {
		if closeErr := engine.Close(); closeErr != nil {
			t.Fatalf("unexpected engine close error: %v", closeErr)
		}
	}()

	output, err := engine.Run(context.Background(), source.NewAnonymous(`
		RETURN WEB::ARTICLE::MARKDOWN("
			<html>
			  <body>
			    <article class='story'>
			      <p>Rendered article bodies should convert to markdown through the session-injected converter for repeated runs across extraction sessions without rebuilding the converter plugin set each time.</p>
			      <p>The article fixture also needs enough readable prose to remain above the meaningful body threshold while still containing a <a href='/story'>Story link</a> that markdown conversion can preserve.</p>
			    </article>
			  </body>
			</html>
		")
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	var markdown string
	if err := json.Unmarshal(output.Content, &markdown); err != nil {
		t.Fatalf("unexpected markdown decode error: %v", err)
	}

	if markdown == "" {
		t.Fatal("expected markdown output")
	}

	if !seenExtractor {
		t.Fatal("expected after-run hook to observe session extractor")
	}
}
