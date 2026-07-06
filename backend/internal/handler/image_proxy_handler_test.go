package handler

import (
	"net/http"
	"net/url"
	"testing"
)

func TestValidatePublicImageURLRejectsPrivateTargets(t *testing.T) {
	tests := []string{
		"",
		"file:///tmp/image.png",
		"http://localhost/image.png",
		"http://app.local/image.png",
		"http://127.0.0.1/image.png",
		"http://10.0.0.1/image.png",
		"http://172.16.0.1/image.png",
		"http://192.168.0.1/image.png",
		"http://[::1]/image.png",
	}

	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := validatePublicImageURL(raw); err == nil {
				t.Fatalf("validatePublicImageURL(%q) returned nil error", raw)
			}
		})
	}
}

func TestValidateImageProxyRedirectRejectsPrivateRedirect(t *testing.T) {
	req := &http.Request{URL: mustParseImageProxyTestURL(t, "http://127.0.0.1/private.png")}
	via := []*http.Request{{URL: mustParseImageProxyTestURL(t, "https://example.com/image.png")}}

	if err := validateImageProxyRedirect(req, via); err == nil {
		t.Fatal("validateImageProxyRedirect returned nil error for private redirect")
	}
}

func TestValidateImageProxyRedirectRejectsRedirectLoop(t *testing.T) {
	req := &http.Request{URL: mustParseImageProxyTestURL(t, "https://example.com/image.png")}
	via := make([]*http.Request, 10)

	if err := validateImageProxyRedirect(req, via); err == nil {
		t.Fatal("validateImageProxyRedirect returned nil error for redirect loop")
	}
}

func mustParseImageProxyTestURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", raw, err)
	}
	return parsed
}
