package onlineconf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const (
	RequestTimeout = 3 * time.Second
	// MaxSnapshotBytes caps the complete uncompressed response, including the envelope.
	MaxSnapshotBytes   = 16 << 20
	maxErrorDrainBytes = 4 << 10
	snapshotPath       = "/internal/onlineconf/snapshot"
)

var (
	ErrSnapshotTooLarge   = errors.New("OnlineConf snapshot exceeds the size limit")
	ErrInvalidSnapshot    = errors.New("invalid OnlineConf snapshot")
	ErrSnapshotConnection = errors.New("OnlineConf snapshot request failed")
)

type StatusError struct {
	StatusCode int
}

func (e StatusError) Error() string {
	return fmt.Sprintf("OnlineConf snapshot returned HTTP %d", e.StatusCode)
}

type Client struct {
	endpoint string
	secret   string
	http     *http.Client
}

func NewClient(baseURL, secret string) (*Client, error) {
	base, err := url.Parse(baseURL)
	if err != nil || base.Hostname() == "" || (base.Scheme != "http" && base.Scheme != "https") ||
		base.User != nil || base.RawQuery != "" || base.ForceQuery || base.Fragment != "" {
		return nil, errors.New("invalid OnlineConf service URL")
	}
	if secret == "" || strings.ContainsFunc(secret, unicode.IsSpace) {
		return nil, errors.New("invalid OnlineConf Bearer secret")
	}
	return &Client{
		endpoint: base.JoinPath(snapshotPath).String(),
		secret:   secret,
		http: &http.Client{
			Timeout:   RequestTimeout,
			Transport: newSnapshotTransport(),
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func newSnapshotTransport() *http.Transport {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: RequestTimeout}).DialContext,
		ForceAttemptHTTP2:   true,
		IdleConnTimeout:     2 * RefreshInterval,
		MaxIdleConnsPerHost: 1,
	}
}

func (c *Client) FetchSnapshot(ctx context.Context) (Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return Snapshot{}, ErrSnapshotConnection
	}
	req.Header.Set("Authorization", "Bearer "+c.secret)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return Snapshot{}, requestError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxErrorDrainBytes))
		return Snapshot{}, StatusError{StatusCode: resp.StatusCode}
	}

	body := &io.LimitedReader{R: resp.Body, N: MaxSnapshotBytes + 1}
	decoder := json.NewDecoder(body)
	var envelope struct {
		Data *snapshotResponse `json:"data"`
	}
	if err := decoder.Decode(&envelope); err != nil {
		return Snapshot{}, decodeSnapshotError(body, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Snapshot{}, decodeSnapshotError(body, err)
	}
	if body.N == 0 {
		return Snapshot{}, ErrSnapshotTooLarge
	}
	if envelope.Data == nil {
		return Snapshot{}, ErrInvalidSnapshot
	}
	return envelope.Data.snapshot()
}

func requestError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded):
		return context.DeadlineExceeded
	default:
		return ErrSnapshotConnection
	}
}

func decodeSnapshotError(body *io.LimitedReader, err error) error {
	if body.N == 0 {
		return ErrSnapshotTooLarge
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return requestError(err)
	}
	return ErrInvalidSnapshot
}
