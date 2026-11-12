package origin

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

func normalize(parsed *url.URL) (origin, error) {
	if parsed == nil || parsed.Opaque != "" || parsed.Host == "" {
		return origin{}, fmt.Errorf("expected an absolute HTTP(S) URL")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return origin{}, fmt.Errorf("scheme must be http or https")
	}

	host := parsed.Hostname()
	if host == "" || strings.ContainsAny(host, "*%\\") {
		return origin{}, fmt.Errorf("invalid hostname")
	}

	if address, parseErr := netip.ParseAddr(host); parseErr == nil {
		if strings.HasPrefix(parsed.Host, "[") && !address.Is6() {
			return origin{}, fmt.Errorf("brackets require an IPv6 address")
		}

		host = address.String()
	} else {
		// IDNA applies its own case mapping to the original labels, as net/http does.
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
		value, err := strconv.ParseUint(text, 10, 16)
		if err != nil {
			return origin{}, fmt.Errorf("invalid port")
		}

		port = value
	} else if strings.HasSuffix(parsed.Host, ":") {
		return origin{}, fmt.Errorf("empty port")
	}

	return origin{scheme: parsed.Scheme, host: host, port: uint16(port)}, nil
}
