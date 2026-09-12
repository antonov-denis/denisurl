package web

import (
	"errors"
	"net/url"
	"strings"
)

var errInvalidTarget = errors.New("target must be a http or https URL")

// normalizeTarget cleans up a user-supplied target and reports whether it is a
// usable absolute http(s) URL.
//
// A missing scheme is assumed to be https, so "example.com" works. The scheme
// is then checked explicitly: url.Parse is permissive and happily accepts
// "javascript:alert(1)" as a URL with scheme "javascript", which would make the
// redirect an XSS vector.
//
// The check for "://" rather than an empty u.Scheme is deliberate too:
// url.Parse("example.com:8080/x") yields scheme "example.com".
func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errInvalidTarget
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	target, err := url.Parse(raw)
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		return "", errInvalidTarget
	}

	return target.String(), nil
}
