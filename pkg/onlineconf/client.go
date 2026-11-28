package onlineconf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const (
	RequestTimeout = 3 * time.Second
	snapshotPath   = "/internal/onlineconf/snapshot"
)

var (
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
			Timeout: RequestTimeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
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
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return Snapshot{}, StatusError{StatusCode: resp.StatusCode}
	}

	decoder := json.NewDecoder(resp.Body)
	var envelope struct {
		Data *snapshotResponse `json:"data"`
	}
	if err := decoder.Decode(&envelope); err != nil {
		return Snapshot{}, decodeError(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Snapshot{}, decodeError(err)
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

func decodeError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return requestError(err)
	}
	return ErrInvalidSnapshot
}
