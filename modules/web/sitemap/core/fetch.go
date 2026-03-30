package core

import (
	"context"
	"io"
	"net/http"
	neturl "net/url"
)

// Fetch downloads and parses a sitemap document.
func Fetch(ctx context.Context, target string, opts Options) (Document, error) {
	body, err := open(ctx, target, opts)
	if err != nil {
		return Document{}, err
	}

	defer body.Close()

	return Parse(ctx, body, target)
}

func open(ctx context.Context, target string, opts Options) (io.ReadCloser, error) {
	parsed, err := neturl.ParseRequestURI(target)
	if err != nil {
		return nil, wrapError(target, StageFetch, err, "invalid sitemap URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, newErrorf(target, StageFetch, "unsupported URL scheme %q", parsed.Scheme)
	}

	if parsed.Host == "" {
		return nil, newError(target, StageFetch, "sitemap URL must include a host")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, wrapError(target, StageFetch, err, "failed to create sitemap request")
	}

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: opts.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, wrapError(target, StageFetch, err, "failed to fetch sitemap")
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_ = resp.Body.Close()

		return nil, newErrorf(target, StageFetch, "unexpected HTTP status %d", resp.StatusCode)
	}

	return resp.Body, nil
}
