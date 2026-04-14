package chttp

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Option func(*HttpClient)

func WithHeader(header map[string]string) Option {
	return func(o *HttpClient) {
		for k, v := range header {
			o.header.Set(k, v)
		}
	}
}

func WithHttpClient(c *http.Client) Option {
	return func(o *HttpClient) {
		o.cli = c
	}
}

func WithAddHeader(key string, value string) Option {
	return func(o *HttpClient) {
		o.header.Add(key, value)
	}
}

func WithContentType(contentType string) Option {
	return func(o *HttpClient) {
		o.header.Set("Content-Type", contentType)
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *HttpClient) {
		if o.cli == nil {
			o.cli = &http.Client{
				Timeout: timeout,
			}
		} else {
			o.cli.Timeout = timeout
		}
	}
}

func WithUserAgent(ua string) Option {
	return func(o *HttpClient) {
		o.header.Set("User-Agent", ua)
	}
}

func WithQuery(query url.Values) Option {
	return func(o *HttpClient) {
		if strings.Contains(o.url, "?") {
			o.url += "&" + query.Encode()
		} else {
			o.url += "?" + query.Encode()
		}
	}
}

func WithBasicAuth(username, password string) Option {
	return func(o *HttpClient) {
		req, _ := http.NewRequest(o.method, o.url, nil)
		req.SetBasicAuth(username, password)
		o.header.Set("Authorization", req.Header.Get("Authorization"))
	}
}
