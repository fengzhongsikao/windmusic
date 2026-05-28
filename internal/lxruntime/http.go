package lxruntime

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type httpClient struct {
	client *http.Client
	vm     *goja.Runtime
	mu     *sync.Mutex
}

func newHTTPClient(vm *goja.Runtime, mu *sync.Mutex) *httpClient {
	return &httpClient{
		client: &http.Client{Timeout: 30 * time.Second},
		vm:     vm,
		mu:     mu,
	}
}

func (h *httpClient) request(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 3 {
		return goja.Undefined()
	}

	reqURL := call.Arguments[0].String()
	options := exportObject(call.Arguments[1])
	callback, ok := goja.AssertFunction(call.Arguments[2])
	if !ok {
		return goja.Undefined()
	}

	ctx, cancel := context.WithCancel(context.Background())
	go h.doRequest(ctx, reqURL, options, callback)

	return h.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		cancel()
		return goja.Undefined()
	})
}

func (h *httpClient) doRequest(ctx context.Context, reqURL string, options map[string]interface{}, callback goja.Callable) {
	respObj, err := h.perform(ctx, reqURL, options)

	h.mu.Lock()
	defer h.mu.Unlock()

	var errVal goja.Value
	if err != nil {
		errVal = h.vm.ToValue(map[string]interface{}{
			"message": err.Error(),
		})
	} else {
		errVal = goja.Null()
	}

	_, _ = callback(goja.Undefined(), errVal, h.vm.ToValue(respObj))
}

func (h *httpClient) perform(ctx context.Context, reqURL string, options map[string]interface{}) (map[string]interface{}, error) {
	startedAt := time.Now()
	method := strings.ToUpper(stringValue(options["method"], "GET"))
	timeout := time.Duration(intValue(options["timeout"], 30)) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client := *h.client
	client.Timeout = timeout

	var body io.Reader
	headers := map[string]string{}
	for key, value := range mapValue(options["headers"]) {
		headers[key] = fmt.Sprint(value)
	}

	switch {
	case options["formData"] != nil:
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		for key, value := range mapValue(options["formData"]) {
			_ = writer.WriteField(key, fmt.Sprint(value))
		}
		_ = writer.Close()
		body = buf
		headers["Content-Type"] = writer.FormDataContentType()
	case options["form"] != nil:
		values := url.Values{}
		for key, value := range mapValue(options["form"]) {
			values.Set(key, fmt.Sprint(value))
		}
		body = strings.NewReader(values.Encode())
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	case options["body"] != nil:
		switch v := options["body"].(type) {
		case string:
			body = strings.NewReader(v)
		case []byte:
			body = bytes.NewReader(v)
		default:
			data, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
			if _, ok := headers["Content-Type"]; !ok {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		log.Printf("[后端:http] 构建请求失败 method=%s url=%s err=%v elapsed=%s", method, reqURL, err, time.Since(startedAt))
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[后端:http] 请求失败 method=%s url=%s err=%v elapsed=%s", method, reqURL, err, time.Since(startedAt))
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("[后端:http] 已收到响应头 method=%s url=%s status=%d elapsed=%s", method, reqURL, resp.StatusCode, time.Since(startedAt))

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[后端:http] 读取响应体失败 method=%s url=%s status=%d err=%v elapsed=%s", method, reqURL, resp.StatusCode, err, time.Since(startedAt))
		return nil, err
	}

	respHeaders := map[string]interface{}{}
	for key, values := range resp.Header {
		if len(values) == 1 {
			respHeaders[key] = values[0]
		} else {
			respHeaders[key] = values
		}
	}

	respObj := map[string]interface{}{
		"statusCode":    resp.StatusCode,
		"statusMessage": resp.Status,
		"headers":       respHeaders,
		"body":          decodeBody(raw, resp.Header.Get("Content-Type")),
	}

	log.Printf("[后端:http] 请求完成 method=%s url=%s status=%d bytes=%d elapsed=%s", method, reqURL, resp.StatusCode, len(raw), time.Since(startedAt))
	return respObj, nil
}

func decodeBody(raw []byte, contentType string) interface{} {
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return ""
	}

	lower := strings.ToLower(contentType)
	if strings.Contains(lower, "json") || (text[0] == '{' || text[0] == '[') {
		var parsed interface{}
		if err := json.Unmarshal(raw, &parsed); err == nil {
			return parsed
		}
	}
	return text
}

func exportObject(value goja.Value) map[string]interface{} {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return map[string]interface{}{}
	}
	exported := value.Export()
	if exported == nil {
		return map[string]interface{}{}
	}
	if obj, ok := exported.(map[string]interface{}); ok {
		return obj
	}
	return map[string]interface{}{}
}

func mapValue(value interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	if obj, ok := value.(map[string]interface{}); ok {
		return obj
	}
	return map[string]interface{}{}
}

func stringValue(value interface{}, fallback string) string {
	if value == nil {
		return fallback
	}
	if s, ok := value.(string); ok && s != "" {
		return s
	}
	return fallback
}

func intValue(value interface{}, fallback int) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return fallback
	}
}

func encodeBuffer(data []byte, format string) string {
	switch strings.ToLower(format) {
	case "hex":
		return hex.EncodeToString(data)
	case "base64":
		return base64.StdEncoding.EncodeToString(data)
	default:
		return string(data)
	}
}

func decodeBufferInput(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("不支持的缓冲区输入")
	}
}
