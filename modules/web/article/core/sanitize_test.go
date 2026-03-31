package core

import (
	"strings"
	"testing"
)

func TestSanitizeHTMLRemovesDangerousAttrsAndURLs(t *testing.T) {
	extractor := NewExtractor()

	safe := extractor.sanitizeHTML(`
		<section class="story-shell" style="display:block">
		  <p onclick="alert(1)">Lead <a href="javascript:alert(2)" title="Bad">Bad link</a></p>
		  <p><a href="/story" data-track="hero">Story</a> <a href="mailto:news@example.com">Email</a> <a href="tel:+15551234567">Call</a></p>
		  <img src="javascript:alert(3)" onerror="steal()" alt="Bad image" />
		  <img src="https://example.com/hero.jpg" onerror="steal()" style="width:100px" alt="Lead image" />
		  <table><tr><th scope="col" onclick="x()">Field</th><td rowspan="2" style="color:red">Value</td></tr></table>
		  <pre><code>const x = 1;</code></pre>
		</section>
	`)
	if safe == nil {
		t.Fatal("expected sanitized html")
	}

	if strings.Contains(*safe, "onclick=") || strings.Contains(*safe, "onerror=") || strings.Contains(*safe, "style=") || strings.Contains(*safe, "class=") || strings.Contains(*safe, "data-track=") {
		t.Fatalf("expected dangerous attrs to be removed, got %q", *safe)
	}

	if strings.Contains(*safe, "javascript:") || strings.Contains(*safe, "data:") {
		t.Fatalf("expected dangerous schemes to be removed, got %q", *safe)
	}

	if !strings.Contains(*safe, `href="/story"`) || !strings.Contains(*safe, `href="mailto:news@example.com"`) || !strings.Contains(*safe, `href="tel:+15551234567"`) {
		t.Fatalf("expected safe links to survive, got %q", *safe)
	}

	if !strings.Contains(*safe, `src="https://example.com/hero.jpg"`) {
		t.Fatalf("expected safe image to survive, got %q", *safe)
	}

	if !strings.Contains(*safe, "<table>") || !strings.Contains(*safe, `scope="col"`) || !strings.Contains(*safe, `rowspan="2"`) || !strings.Contains(*safe, "<pre><code>const x = 1;</code></pre>") {
		t.Fatalf("expected formatting tags to survive, got %q", *safe)
	}
}
