package newznab

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/schema"
	"github.com/henges/newznab-proxy/xmlutil"
)

const ua = "newznab-client/0.0.1"

type Client struct {
	cl        *http.Client
	baseURL   string
	apiKey    string
	userAgent string
}

type clientOptions struct {
	userAgent string
}

type ClientOption func(options *clientOptions)

func WithUserAgent(s string) ClientOption {
	return func(options *clientOptions) {
		options.userAgent = s
	}
}

var encoderOnce sync.Once
var encoder *schema.Encoder

func NewClient(baseURL, apiKey string, opts ...ClientOption) *Client {
	options := &clientOptions{
		userAgent: ua,
	}
	for _, o := range opts {
		o(options)
	}

	return &Client{
		cl:        &http.Client{},
		baseURL:   baseURL,
		apiKey:    apiKey,
		userAgent: options.userAgent,
	}
}

func getEncoder() *schema.Encoder {

	encoderOnce.Do(func() {
		encoder = schema.NewEncoder()
	})
	return encoder
}

func (c *Client) Search(ctx context.Context, params SearchParams) (*RssFeed, error) {

	v := make(url.Values)
	err := getEncoder().Encode(params, v)
	if err != nil {
		return nil, err
	}
	v.Set("t", "search")
	v.Set("apikey", c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var ret RssFeed
	err = xmlutil.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
