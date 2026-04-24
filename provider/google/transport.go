package google

import "net/http"

type zenMuxTransport struct {
	apiKey string
	base   http.RoundTripper
}

func newTransport(apiKey string, base http.RoundTripper) *zenMuxTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &zenMuxTransport{apiKey: apiKey, base: base}
}

func (t *zenMuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("Authorization", "Bearer "+t.apiKey)
	r.Header.Del("x-goog-api-key")
	return t.base.RoundTrip(r)
}
