package core

import "testing"

func TestMatch(t *testing.T) {
	t.Run("uses exact user-agent groups before wildcard fallback", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/blocked"}},
				{UserAgents: []string{"FerretBot"}, Allow: []string{"/blocked"}},
			},
		}

		result := Match(doc, "/blocked/page", "FerretBot")
		if !result.Allowed {
			t.Fatalf("expected allowed result, got %+v", result)
		}

		if result.Directive == nil || *result.Directive != directiveAllow {
			t.Fatalf("unexpected directive %v", result.Directive)
		}

		if result.UserAgent != "FerretBot" {
			t.Fatalf("expected exact-match userAgent %q, got %q", "FerretBot", result.UserAgent)
		}
	})

	t.Run("falls back to wildcard groups", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/blocked"}},
			},
		}

		result := Match(doc, "/blocked/page", "OtherBot")
		if result.Allowed {
			t.Fatalf("expected disallowed result, got %+v", result)
		}

		if result.UserAgent != "*" {
			t.Fatalf("expected wildcard fallback userAgent %q, got %q", "*", result.UserAgent)
		}
	})

	t.Run("merges repeated matching groups", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"FerretBot"}, Disallow: []string{"/private"}},
				{UserAgents: []string{"FerretBot"}, Allow: []string{"/private/public"}},
			},
		}

		result := Match(doc, "/private/public/page", "FerretBot")
		if !result.Allowed {
			t.Fatalf("expected allowed result, got %+v", result)
		}

		if result.Pattern == nil || *result.Pattern != "/private/public" {
			t.Fatalf("unexpected pattern %v", result.Pattern)
		}
	})

	t.Run("prefers the longest matching rule", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Allow:      []string{"/private/public"},
					Disallow:   []string{"/private"},
				},
			},
		}

		result := Match(doc, "/private/public/page", "Crawler")
		if !result.Allowed {
			t.Fatalf("expected allowed result, got %+v", result)
		}
	})

	t.Run("allow wins equal-specificity ties", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Allow:      []string{"/same"},
					Disallow:   []string{"/same"},
				},
			},
		}

		result := Match(doc, "/same/path", "Crawler")
		if !result.Allowed {
			t.Fatalf("expected allowed result, got %+v", result)
		}
	})

	t.Run("supports wildcard matching", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/private/*/download"}},
			},
		}

		result := Match(doc, "/private/alpha/download/file", "Crawler")
		if result.Allowed {
			t.Fatalf("expected disallowed result, got %+v", result)
		}
	})

	t.Run("supports end anchors", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/exact$"}},
			},
		}

		if result := Match(doc, "/exact", "Crawler"); result.Allowed {
			t.Fatalf("expected exact path to be disallowed, got %+v", result)
		}

		if result := Match(doc, "/exact/more", "Crawler"); !result.Allowed {
			t.Fatalf("expected anchored mismatch to be allowed, got %+v", result)
		}
	})

	t.Run("allows unmatched paths by default", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/private"}},
			},
		}

		result := Match(doc, "/public", "Crawler")
		if !result.Allowed || result.Directive != nil || result.Pattern != nil {
			t.Fatalf("expected default-allow result, got %+v", result)
		}

		if result.UserAgent != "*" {
			t.Fatalf("expected wildcard fallback userAgent %q, got %q", "*", result.UserAgent)
		}
	})

	t.Run("always allows robots.txt", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/"}},
			},
		}

		if !Allows(doc, "/robots.txt", "Crawler") {
			t.Fatal("expected /robots.txt to be allowed")
		}

		if !Allows(doc, "/robots.txt?cache=1", "Crawler") {
			t.Fatal("expected /robots.txt with query to be allowed")
		}

		result := Match(doc, "/robots.txt", "Crawler")
		if result.UserAgent != "Crawler" {
			t.Fatalf("expected implicit-allow userAgent %q, got %q", "Crawler", result.UserAgent)
		}
	})

	t.Run("no matching groups keep the requested token", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"SpecificBot"}, Disallow: []string{"/private"}},
			},
		}

		result := Match(doc, "/public", "Crawler")
		if !result.Allowed {
			t.Fatalf("expected default-allow result, got %+v", result)
		}

		if result.UserAgent != "Crawler" {
			t.Fatalf("expected unmatched userAgent %q, got %q", "Crawler", result.UserAgent)
		}
	})

	t.Run("matches user-agent case-insensitively", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"FerretBot"}, Disallow: []string{"/private"}},
			},
		}

		result := Match(doc, "/private", "ferretbot")
		if result.Allowed {
			t.Fatalf("expected case-insensitive match to disallow, got %+v", result)
		}
	})

	t.Run("normalizes representative percent-encoding cases", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Disallow: []string{
						"/foo/bar/%62%61%7A",
						"/foo/bar/%E3%83%84",
					},
				},
			},
		}

		if result := Match(doc, "/foo/bar/baz", "Crawler"); result.Allowed {
			t.Fatalf("expected encoded ascii rule to match, got %+v", result)
		}

		if result := Match(doc, "/foo/bar/ツ", "Crawler"); result.Allowed {
			t.Fatalf("expected encoded utf8 rule to match, got %+v", result)
		}
	})

	t.Run("supports repeated wildcards and empty segments", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/**/foo"}},
			},
		}

		if result := Match(doc, "/abc/def/foo", "Crawler"); result.Allowed {
			t.Fatalf("expected repeated wildcard rule to match, got %+v", result)
		}
	})

	t.Run("supports wildcard at start middle and end", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Disallow: []string{
						"/*suffix",
						"/mid/*/tail",
						"/prefix*",
					},
				},
			},
		}

		if result := Match(doc, "/aaa/suffix", "Crawler"); result.Allowed {
			t.Fatalf("expected wildcard-at-start rule to match, got %+v", result)
		}

		if result := Match(doc, "/mid/value/tail/rest", "Crawler"); result.Allowed {
			t.Fatalf("expected wildcard-in-middle rule to match, got %+v", result)
		}

		if result := Match(doc, "/prefix-and-more", "Crawler"); result.Allowed {
			t.Fatalf("expected wildcard-at-end rule to match, got %+v", result)
		}
	})

	t.Run("supports anchored wildcard patterns", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{UserAgents: []string{"*"}, Disallow: []string{"/foo/*/bar$"}},
			},
		}

		if result := Match(doc, "/foo/value/bar", "Crawler"); result.Allowed {
			t.Fatalf("expected anchored wildcard rule to match, got %+v", result)
		}

		if result := Match(doc, "/foo/value/bar/baz", "Crawler"); !result.Allowed {
			t.Fatalf("expected anchored wildcard mismatch to allow, got %+v", result)
		}
	})

	t.Run("anchored wildcard matches when literal appears multiple times", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Disallow: []string{
						"/*.json$",
						"/*/a$",
					},
				},
			},
		}

		if result := Match(doc, "/api/data.json.json", "Crawler"); result.Allowed {
			t.Fatalf("expected /*.json$ to match path with repeated .json, got %+v", result)
		}

		if result := Match(doc, "/x/a/a", "Crawler"); result.Allowed {
			t.Fatalf("expected /*/a$ to match path with repeated /a, got %+v", result)
		}

		if result := Match(doc, "/api/data.json.txt", "Crawler"); !result.Allowed {
			t.Fatalf("expected /*.json$ to not match non-.json suffix, got %+v", result)
		}
	})

	t.Run("trailing wildcard with end anchor matches any path", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Disallow: []string{
						"/*$",
						"/foo/*$",
					},
				},
			},
		}

		if result := Match(doc, "/any/path/here", "Crawler"); result.Allowed {
			t.Fatalf("expected /*$ to match any path, got %+v", result)
		}

		if result := Match(doc, "/foo/bar/baz", "Crawler"); result.Allowed {
			t.Fatalf("expected /foo/*$ to match path under /foo/, got %+v", result)
		}
	})

	t.Run("treats regex significant characters as plain literals", func(t *testing.T) {
		doc := Document{
			Groups: []Group{
				{
					UserAgents: []string{"*"},
					Disallow: []string{
						"/file?.txt",
						"/report.(final)",
					},
				},
			},
		}

		if result := Match(doc, "/file?.txt", "Crawler"); result.Allowed {
			t.Fatalf("expected literal question mark rule to match, got %+v", result)
		}

		if result := Match(doc, "/report.(final)/v2", "Crawler"); result.Allowed {
			t.Fatalf("expected literal punctuation rule to match, got %+v", result)
		}
	})
}
