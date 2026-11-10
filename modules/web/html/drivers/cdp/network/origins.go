package network

import (
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

type origin struct {
	scheme string
	host   string
	port   uint16
}

var originNames = idna.New(idna.MapForLookup(), idna.StrictDomainName(true), idna.ValidateLabels(true), idna.VerifyDNSLength(true))

func compileOrigins(values []string) (map[origin]struct{}, error) {
	origins := make(map[origin]struct{}, len(values))

	for index, value := range values {
		parsed, err := parseOrigin(value, true)
		if err != nil {
			// Configuration may contain credentials; report the position without
			// echoing the supplied URL or a parser error containing it.
			return nil, fmt.Errorf("allowed origin %d: %w", index+1, err)
		}

		origins[parsed] = struct{}{}
	}

	return origins, nil
}

func parseOrigin(value string, configuration bool) (origin, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Opaque != "" || parsed.Host == "" {
		return origin{}, fmt.Errorf("expected an absolute HTTP(S) URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return origin{}, fmt.Errorf("scheme must be http or https")
	}

	if configuration && (parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") ||
		parsed.RawPath != "" || parsed.ForceQuery || parsed.RawQuery != "" || strings.Contains(value, "#")) {
		return origin{}, fmt.Errorf("expected an origin without credentials, path, query, or fragment")
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" || strings.ContainsAny(host, "*%\\") {
		return origin{}, fmt.Errorf("invalid hostname")
	}

	if address, parseErr := netip.ParseAddr(host); parseErr == nil {
		if strings.HasPrefix(parsed.Host, "[") && !address.Is6() {
			return origin{}, fmt.Errorf("brackets require an IPv6 address")
		}

		host = address.String()
	} else {
		ascii, nameErr := originNames.ToASCII(host)
		if nameErr != nil || strings.HasPrefix(parsed.Host, "[") || strings.ContainsAny(host, ":[]") {
			return origin{}, fmt.Errorf("invalid hostname")
		}

		host = ascii
	}

	port := uint64(80)
	if parsed.Scheme == "https" {
		port = 443
	}

	if text := parsed.Port(); text != "" {
		port, err = strconv.ParseUint(text, 10, 16)
		if err != nil {
			return origin{}, fmt.Errorf("invalid port")
		}
	} else if strings.HasSuffix(parsed.Host, ":") {
		return origin{}, fmt.Errorf("empty port")
	}

	return origin{scheme: parsed.Scheme, host: host, port: uint16(port)}, nil
}
