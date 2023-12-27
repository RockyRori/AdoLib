package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

//go:generate mockgen -package mock -source ./http_client.go -destination ./mock/mock_http_client.go

// HTTPClient HTTP客户端服务接口。
type HTTPClient interface {
	Get(ctx context.Context, url string, queryValues url.Values, headers map[string]string) (respCode int, respData interface{}, err error)
	GetNoUnmarshal(ctx context.Context, url string, queryValues url.Values, headers map[string]string) (respCode int, respBody []byte, err error)
	Delete(ctx context.Context, url string, headers map[string]string) (respCode int, respData interface{}, err error)
	DeleteNoUnmarshal(ctx context.Context, url string, headers map[string]string) (respCode int, respBody []byte, err error)
	Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PostNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PutNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	Patch(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PatchNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
}

// httpClient HTTP客户端结构。
type httpClient struct {
	client *http.Client
}

// HttpClientOptions httpClient 配置信息。
type HttpClientOptions struct {
	TimeOut int
}

// NewRawHTTPClient 创建原生HTTP客户端对象。
func NewRawHTTPClient() *http.Client {
	opts := HttpClientOptions{
		TimeOut: 10,
	}
	return NewRawHTTPClientWithOptions(opts)
}

// NewHTTPClientWithOptions 根据配置创建HTTP客户端对象。
func NewHTTPClientWithOptions(opts HttpClientOptions) HTTPClient {
	client := &httpClient{
		client: NewRawHTTPClientWithOptions(opts),
	}

	return client
}

// NewRawHTTPClientWithOptions 根据配置创建原生HTTP客户端对象。
func NewRawHTTPClientWithOptions(opts HttpClientOptions) *http.Client {
	rawClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost:   100,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Duration(opts.TimeOut) * time.Second,
	}

	return rawClient
}

// NewHTTPClient 创建HTTP客户端对象。
func NewHTTPClient() HTTPClient {
	client := &httpClient{
		client: NewRawHTTPClient(),
	}

	return client
}

// Get 返回序列化对象。
func (c *httpClient) Get(ctx context.Context, rawURL string, queryValues url.Values, headers map[string]string) (respCode int, respData interface{}, err error) {
	url, err := c.generateURL(rawURL, queryValues)
	if err != nil {
		log.Println(err.Error())
		return
	}

	return c.httpDo(ctx, http.MethodGet, url.String(), headers, nil)
}

// GetNoUnmarshal 返回text。
func (c *httpClient) GetNoUnmarshal(ctx context.Context, rawURL string, queryValues url.Values, headers map[string]string) (respCode int, respBody []byte, err error) {
	url, err := c.generateURL(rawURL, queryValues)
	if err != nil {
		log.Println(err.Error())
		return
	}

	return c.httpDoNoUnmarshal(ctx, http.MethodGet, url.String(), headers, nil)
}

// Post 传入序列化对象，返回序列化对象。
func (c *httpClient) Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodPost, url, headers, reqParam)
}

// PostNoUnmarshal 传入序列化对象，返回text。
func (c *httpClient) PostNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodPost, url, headers, reqParam)
}

// Put 传入序列化对象，返回序列化对象。
func (c *httpClient) Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodPut, url, headers, reqParam)
}

// PutNoUnmarshal 传入序列化对象，返回text。
func (c *httpClient) PutNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodPut, url, headers, reqParam)
}

// Delete 返回序列化对象。
func (c *httpClient) Delete(ctx context.Context, url string, headers map[string]string) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodDelete, url, headers, nil)
}

// DeleteNoUnmarshal 传入序列化对象，返回text。
func (c *httpClient) DeleteNoUnmarshal(ctx context.Context, url string, headers map[string]string) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodDelete, url, headers, nil)
}

// Patch 传入序列化对象，返回序列化对象。
func (c *httpClient) Patch(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error) {
	return c.httpDo(ctx, http.MethodPatch, url, headers, reqParam)
}

// PatchNoUnmarshal 传入序列化对象，返回text。
func (c *httpClient) PatchNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error) {
	return c.httpDoNoUnmarshal(ctx, http.MethodPatch, url, headers, reqParam)
}

// 反序列化返回内容。
func (c *httpClient) httpDo(ctx context.Context, mtehod string, url string, headers map[string]string,
	reqParam interface{}) (respCode int, respData interface{}, err error) {

	respCode, respBody, err := c.httpDoNoUnmarshal(ctx, mtehod, url, headers, reqParam)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if len(respBody) == 0 {
		return
	}

	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		log.Println(err.Error())
	}
	return
}

// 返回原始respBody, 不进行反序列化。
func (c *httpClient) httpDoNoUnmarshal(ctx context.Context, mtehod string, url string, headers map[string]string,
	reqParam interface{}) (respCode int, respBody []byte, err error) {

	if c.client == nil {
		return 0, nil, errors.New("http client is unavailable")
	}

	req, err := c.generateReq(ctx, mtehod, url, headers, reqParam)
	if err != nil {
		log.Println(err.Error())
		return 0, nil, err
	}

	// 把 trace 上下文注入到请求的 header 中
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.Println(closeErr.Error())
		}
	}()
	respBody, err = io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	return
}

func (c *httpClient) generateURL(rawURL string, queryValues url.Values) (*url.URL, error) {
	uri, err := url.Parse(rawURL)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	if queryValues != nil {
		values := uri.Query()
		for k, v := range values {
			queryValues[k] = v
		}
		uri.RawQuery = queryValues.Encode()
	}

	return uri, err
}

func (c *httpClient) generateReq(ctx context.Context, httpMethod string, url string, headers map[string]string, reqParam interface{}) (req *http.Request, err error) {
	if reqParam != nil {
		var reader *bytes.Reader
		if v, ok := reqParam.([]byte); ok {
			reader = bytes.NewReader(v)
		} else {
			reqData, err := json.Marshal(reqParam)
			if err != nil {
				log.Println(err.Error())
				return nil, err
			}
			reader = bytes.NewReader(reqData)
		}
		req, err = http.NewRequestWithContext(ctx, httpMethod, url, reader)
	} else {
		req, err = http.NewRequestWithContext(ctx, httpMethod, url, nil)
	}

	if err != nil {
		log.Println(err.Error())
		return
	}

	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
	return
}
