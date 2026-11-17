package platform

import (
	"errors"
	"testing"
)

func TestOpenURLRejectsUnsafeSchemesBeforePlatformDispatch(t *testing.T) {
	for _, raw := range []string{
		"javascript:alert(1)",
		"file:///etc/passwd",
		"https://user@example.test/",
		"https:///missing-host",
		"https://example.test/\nmalicious",
	} {
		if err := OpenURL(raw); !errors.Is(err, ErrUnsupportedURL) {
			t.Errorf("OpenURL(%q) error = %v, want ErrUnsupportedURL", raw, err)
		}
	}
}

func TestExternalURLAllowsWebAndMailLinks(t *testing.T) {
	for _, raw := range []string{"https://example.test/path", "http://example.test/", "mailto:waycord@example.test"} {
		if err := validateExternalURL(raw); err != nil {
			t.Errorf("validateExternalURL(%q) error = %v", raw, err)
		}
	}
}
