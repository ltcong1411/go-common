package http

import (
	"context"
	"errors"

	"github.com/ltcong1411/go-common/logging"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client provide a simple http service which already including the retry and sending metrics,
// It's include ready to use GET/POST methods and general func Do will suitable any other use-cases
// Default included headers Content-Type, Accept, X-Request-Id, and ability to add more if needed to
//
//go:generate mockery --name=Client --case=snake --disable-version-string
type Client interface {
	Do(ctx context.Context, method, endpointName, reqURL string, req []byte, resp interface{}, headers map[string]string, opts ...Options) (int, error)
	GET(ctx context.Context, endpointName, reqURL string, resp interface{}, headers map[string]string, opts ...Options) (int, error)
	POST(ctx context.Context, endpointName, reqURL string, req []byte, resp interface{}, headers map[string]string, opts ...Options) (int, error)
	PUT(ctx context.Context, endpointName, reqURL string, req []byte, resp interface{}, headers map[string]string, opts ...Options) (int, error)
}

type httpClientImpl struct {
	cfg    *TransportConfig
	client *http.Client
}

func NewHTTPClient(cfg *TransportConfig) Client {
	return &httpClientImpl{
		cfg:    cfg,
		client: &http.Client{},
	}
}

func (t *httpClientImpl) Do(ctx context.Context, method, endpointName string, reqURL string, req []byte, resp interface{}, headers map[string]string, opts ...Options) (responseCode int, err error) {
	requestBody := strings.NewReader(string(req))
	var httpReq *http.Request
	httpReq, err = http.NewRequest(method, reqURL, requestBody)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if err := t.retry(ctx, endpointName, method, func(ctx context.Context) (int, error) {
		responseCode, err = t.sendRequest(ctx, endpointName, httpReq, &resp, headers)
		return responseCode, err
	}, opts...); err != nil {
		logger := logging.FromContext(ctx)
		logger.Errorw(fmt.Sprintf("[%s] %s failed with status code %d", method, endpointName, responseCode), "err", err)
		return responseCode, err
	}
	return responseCode, err
}

func (t *httpClientImpl) GET(ctx context.Context, endpointName string, reqURL string, resp interface{}, headers map[string]string, opts ...Options) (responseCode int, err error) {
	var httpReq *http.Request
	httpReq, err = http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if err := t.retry(ctx, endpointName, http.MethodGet, func(ctx context.Context) (int, error) {
		responseCode, err = t.sendRequest(ctx, endpointName, httpReq, &resp, headers)
		return responseCode, err
	}, opts...); err != nil {
		logger := logging.FromContext(ctx)
		logger.Errorw(fmt.Sprintf("[GET] %s failed with status code %d", endpointName, responseCode), "err", err)
		return responseCode, err
	}
	return responseCode, nil
}

func (t *httpClientImpl) POST(ctx context.Context, endpointName string, reqURL string, req []byte, resp interface{}, headers map[string]string, opts ...Options) (responseCode int, err error) {
	requestBody := strings.NewReader(string(req))
	var httpReq *http.Request
	httpReq, err = http.NewRequest(http.MethodPost, reqURL, requestBody)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if err := t.retry(ctx, endpointName, http.MethodPost, func(ctx context.Context) (int, error) {
		responseCode, err = t.sendRequest(ctx, endpointName, httpReq, &resp, headers)
		return responseCode, err
	}, opts...); err != nil {
		logger := logging.FromContext(ctx)
		logger.Errorw(fmt.Sprintf("[POST] %s failed with status code %d", endpointName, responseCode), "err", err)
		return responseCode, err
	}
	return responseCode, nil
}

func (t *httpClientImpl) PUT(ctx context.Context, endpointName string, reqURL string, req []byte, resp interface{}, headers map[string]string, opts ...Options) (responseCode int, err error) {
	requestBody := strings.NewReader(string(req))
	var httpReq *http.Request
	httpReq, err = http.NewRequest(http.MethodPut, reqURL, requestBody)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	logger := logging.FromContext(ctx)
	if err := t.retry(ctx, endpointName, http.MethodPut, func(ctx context.Context) (int, error) {
		responseCode, err = t.sendRequest(ctx, endpointName, httpReq, &resp, headers)
		logger.Debugw(fmt.Sprintf("[%s]-[%d] %s %s", httpReq.Method, responseCode, httpReq.URL.String(), string(req)))
		return responseCode, err
	}, opts...); err != nil {
		logger.Errorw(fmt.Sprintf("[PUT] %s failed with status code %d", endpointName, responseCode), "err", err)
		return responseCode, err
	}
	return responseCode, nil
}

func (t *httpClientImpl) sendRequest(ctx context.Context, endpointName string, req *http.Request, resp interface{}, headers map[string]string) (httpRespCode int, err error) {

	// Set default header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set custom header
	if len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	req = req.WithContext(ctx)

	logger := logging.FromContext(ctx)

	httpResp, err := t.client.Do(req)
	if err != nil {
		logger.Errorw(fmt.Sprintf("[%s] %s failed", req.Method, endpointName), "err", err)
		return http.StatusInternalServerError, errors.New(
			fmt.Sprintf("[%s] %s failed - err: %v", req.Method, endpointName, err))
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	httpRespCode = httpResp.StatusCode
	httpRespBody, _ := io.ReadAll(httpResp.Body)

	if httpRespCode < http.StatusOK || httpRespCode >= http.StatusBadRequest {
		var errRes error
		var errBody *errorResponse

		if err = json.Unmarshal(httpRespBody, &errBody); err != nil {
			errRes = errors.New(
				fmt.Sprintf("[%s] %s got unexpected error code %d - response: %s", req.Method, endpointName, httpRespCode, errBody))
		} else {
			errRes = errors.New(fmt.Sprintf("[%s] %s got unexpected error code %d", req.Method, endpointName, httpRespCode))
		}
		logger.Errorw(fmt.Sprintf("[%s] %s got unexpected error code", req.Method, endpointName), "err", errRes, "request_url", httpResp.Request.URL, "response_code", httpRespCode, "httpRespBody", string(httpRespBody))
		return httpRespCode, errRes
	}

	// Not parsing body in-case request have no content
	if httpRespCode != http.StatusNoContent && len(httpRespBody) != 0 {
		if err = json.Unmarshal(httpRespBody, &resp); err != nil {
			logger.Errorw(fmt.Sprintf("[%s] %s failed, unable to decode repose body", req.Method, endpointName), "err", err, "request_url", httpResp.Request.URL, "response_code", httpRespCode)
			return httpRespCode, errors.New(fmt.Sprintf("[%s] %s failed, unable to decode repose body", req.Method, endpointName))
		}
	}

	logger.Infow(fmt.Sprintf("[%s] %s success", req.Method, endpointName), "request_url", httpResp.Request.URL, "response_code", httpRespCode)
	return httpRespCode, nil
}

// errorResponse is a general error struct to parsing to get more error on the http response body
type errorResponse struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

func (e *errorResponse) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s| %s", e.Err, e.Message)
}
