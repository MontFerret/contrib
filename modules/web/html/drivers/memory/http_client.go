package memory

import (
	"net/http"
	"net/url"

	"github.com/sethgrid/pester"
)

func newHTTPClient(options *Options) (httpClient *pester.Client) {
	httpClient = pester.New()
	httpClient.Concurrency = options.Concurrency
	httpClient.MaxRetries = options.MaxRetries
	httpClient.Backoff = options.Backoff
	httpClient.Timeout = options.Timeout

	if options.HTTPTransport != nil {
		httpClient.Transport = options.HTTPTransport
	}

	if options.Proxy == "" {
		return
	}

	if err := addProxy(httpClient, options.Proxy); err != nil {
		return
	}

	return
}

func addProxy(httpClient *pester.Client, proxyStr string) error {
	if proxyStr == "" {
		return nil
	}

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return err
	}

	proxy := http.ProxyURL(proxyURL)

	if httpClient.Transport != nil {
		httpClient.Transport.(*http.Transport).Proxy = proxy

		return nil
	}

	httpClient.Transport = &http.Transport{Proxy: proxy}

	return nil
}
