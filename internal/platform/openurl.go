package platform

import (
	"errors"
	"net/url"
	"strings"
	"unicode"
)

var ErrUnsupportedURL = errors.New("unsupported URL scheme")

// OpenURL accepts only explicit web and mail links from user-initiated UI actions.
func OpenURL(raw string) error {
	if err := validateExternalURL(raw); err != nil {
		return err
	}
	return openURL(raw)
}

func validateExternalURL(raw string) error {
	for _, r := range raw {
		if unicode.IsControl(r) {
			return ErrUnsupportedURL
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ErrUnsupportedURL
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		if u.Host == "" || u.User != nil {
			return ErrUnsupportedURL
		}
	case "mailto":
		if u.Opaque == "" && u.Path == "" {
			return ErrUnsupportedURL
		}
	default:
		return ErrUnsupportedURL
	}
	return nil
}
