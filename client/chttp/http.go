package chttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xslasd/goxf/log"
)

var defaultClient = &http.Client{
	Timeout: 30 * time.Second,
}

type HttpClient struct {
	method string
	url    string
	body   io.Reader
	header http.Header
	ctx    context.Context

	cli *http.Client
}

type Response struct {
	*http.Response
	Body []byte
}

func (r *Response) Unmarshal(v any) error {
	return json.Unmarshal(r.Body, v)
}

func (r *Response) String() string {
	return string(r.Body)
}

func (c *HttpClient) send() (*Response, error) {
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(c.ctx, c.method, c.url, c.body)
	if err != nil {
		return nil, err
	}

	for k, v := range c.header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	cli := c.cli
	if cli == nil {
		cli = defaultClient
	}

	resp, err := cli.Do(req)
	if err != nil {
		log.Error("chttp request failed", log.String("method", c.method), log.String("url", c.url), log.FieldErr(err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("chttp read body failed", log.String("method", c.method), log.String("url", c.url), log.FieldErr(err))
		return nil, err
	}

	log.Debug("chttp request ok",
		log.String("method", c.method),
		log.String("url", c.url),
		log.Int("status", resp.StatusCode),
		log.Int("size", len(body)),
	)

	return &Response{
		Response: resp,
		Body:     body,
	}, nil
}

func Get(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, "GET", url, nil, opts...)
}

func Post(ctx context.Context, url string, body io.Reader, opts ...Option) (*Response, error) {
	return Do(ctx, "POST", url, body, opts...)
}

func PostJSON(ctx context.Context, url string, body any, opts ...Option) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithContentType("application/json; charset=utf-8"))
	opts = append(opts, WithAddHeader("Accept", "application/json"))
	return Do(ctx, "POST", url, bytes.NewReader(data), opts...)
}

func PostForm(ctx context.Context, urlStr string, data url.Values, opts ...Option) (*Response, error) {
	opts = append(opts, WithContentType("application/x-www-form-urlencoded"))
	return Do(ctx, "POST", urlStr, strings.NewReader(data.Encode()), opts...)
}

func Put(ctx context.Context, url string, body io.Reader, opts ...Option) (*Response, error) {
	return Do(ctx, "PUT", url, body, opts...)
}

func Delete(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, "DELETE", url, nil, opts...)
}

func Head(ctx context.Context, url string, opts ...Option) (*Response, error) {
	return Do(ctx, "HEAD", url, nil, opts...)
}

func Patch(ctx context.Context, url string, body io.Reader, opts ...Option) (*Response, error) {
	return Do(ctx, "PATCH", url, body, opts...)
}

func Do(ctx context.Context, method, url string, body io.Reader, opts ...Option) (*Response, error) {
	client := &HttpClient{
		ctx:    ctx,
		method: method,
		url:    url,
		body:   body,
		header: make(http.Header),
	}
	for _, o := range opts {
		o(client)
	}
	return client.send()
}
