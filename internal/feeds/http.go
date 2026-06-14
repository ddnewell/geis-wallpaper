package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// userAgent is sent on every request; NOAA/NASA endpoints expect a descriptive
// UA and may reject blank ones.
const userAgent = "geis-wallpaper/1.0 (+https://github.com/ddnewell/geis-wallpaper)"

// Client is the shared HTTP client. The overall timeout is the single most
// important guard against a hung upstream wedging the fetcher. We also raise
// the TLS-handshake timeout above Go's 10s default: some endpoints (e.g.
// wheretheiss.at) negotiate slowly on constrained networks, and the default
// would abort an otherwise-fine request.
var Client = &http.Client{
	Timeout:   25 * time.Second,
	Transport: newTransport(),
}

func newTransport() *http.Transport {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Transport{TLSHandshakeTimeout: 20 * time.Second}
	}
	tc := t.Clone()
	tc.TLSHandshakeTimeout = 20 * time.Second
	return tc
}

// GetBytes performs a GET and returns the body, enforcing a non-error status.
func GetBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64<<20)) // 64 MiB cap
}

// GetJSON performs a GET and decodes the JSON body into v.
func GetJSON(ctx context.Context, url string, v any) error {
	b, err := GetBytes(ctx, url)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}
