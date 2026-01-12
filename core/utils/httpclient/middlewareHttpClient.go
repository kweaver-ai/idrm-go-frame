package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

// HTTPError 服务错误结构体
type HTTPError struct {
	Cause   string                 `json:"cause"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Detail  map[string]interface{} `json:"detail,omitempty"`
}

func (err HTTPError) Error() string {
	errstr, _ := jsoniter.Marshal(err)
	return string(errstr)
}

// ExHTTPError 其他服务响应的错误结构体
type ExHTTPError struct {
	Status int
	Body   []byte
}

func (err ExHTTPError) Error() string {
	return string(err.Body)
}

// HTTPClient HTTP客户端服务接口
type HTTPClient interface {
	Get(ctx context.Context, url string, headers map[string]string) (respParam interface{}, err error)
	Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error)
	Delete(ctx context.Context, url string, headers map[string]string) (respParam interface{}, err error)
}

var (
	httpOnce sync.Once
	client   HTTPClient
)

// httpClient HTTP客户端结构
type httpClient struct {
	client *http.Client
}

// NewHTTPClient 创建HTTP客户端对象
func NewMiddlewareHTTPClient(hc *http.Client) HTTPClient {
	httpOnce.Do(func() {
		client = &httpClient{
			client: hc,
		}
	})

	return client
}

// Get http client get
func (c *httpClient) Get(ctx context.Context, url string, headers map[string]string) (respParam interface{}, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	_, respParam, err = c.httpDo(req, headers)
	return
}

// Post http client post
func (c *httpClient) Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	var reqBody []byte
	if v, ok := reqParam.([]byte); ok {
		reqBody = v
	} else {
		reqBody, err = jsoniter.Marshal(reqParam)
		if err != nil {
			return
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Put http client put
func (c *httpClient) Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respParam interface{}, err error) {
	reqBody, err := jsoniter.Marshal(reqParam)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(reqBody))
	if err != nil {
		return
	}

	respCode, respParam, err = c.httpDo(req, headers)
	return
}

// Delete http client delete
func (c *httpClient) Delete(ctx context.Context, url string, headers map[string]string) (respParam interface{}, err error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return
	}

	_, respParam, err = c.httpDo(req, headers)
	return
}

func (c *httpClient) httpDo(req *http.Request, headers map[string]string) (respCode int, respParam interface{}, err error) {
	if c.client == nil {
		return 0, nil, errors.New("http client is unavailable")
	}

	c.addHeaders(req, headers)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {

		}
	}()
	body, err := io.ReadAll(resp.Body)
	respCode = resp.StatusCode
	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		httpErr := HTTPError{}
		err = jsoniter.Unmarshal(body, &httpErr)
		if err != nil {
			// Unmarshal失败时转成内部错误, body为空Unmarshal失败
			err = fmt.Errorf("code:%v,header:%v,body:%v", respCode, resp.Header, string(body))
		} else {
			err = ExHTTPError{
				Body:   body,
				Status: respCode,
			}
		}
		return
	}

	if len(body) != 0 {
		err = jsoniter.Unmarshal(body, &respParam)
	}

	return
}

func (c *httpClient) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		if len(v) > 0 {
			req.Header.Add(k, v)
		}
	}
}
