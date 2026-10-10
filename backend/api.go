package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func setDefaultHeaders(header http.Header) {
	header.Set("accept", "*/*")
	header.Set("accept-language", "zh-CN,zh;q=0.9")
	header.Set("origin", "https://www.bilibili.com/")
	header.Set("referer", "https://www.bilibili.com/")
	header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36")
}

type Request struct {
	Ctx     context.Context `json:"-"`
	Tx      string          `json:"tx"`
	Method  string          `json:"method"`
	URL     string          `json:"url"`
	CSRF    string          `json:"csrf"`
	Headers [][]string      `json:"headers"`
	Body    json.RawMessage `json:"body"`
	Timeout int64           `json:"timeout"` // in milliseconds
}

type Response struct {
	Tx         string          `json:"tx"`
	StatusCode int             `json:"status_code"`
	Status     string          `json:"status"`
	Headers    [][]string      `json:"headers"`
	Body       json.RawMessage `json:"body"`
}

func ProxyAPI(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tx := NewTx()
	log := NewLogger("ProxyAPI").Tx(tx)

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if json.Valid(reqBody) {
		log.Req(json.RawMessage(reqBody))
	} else {
		log.Req(string(reqBody))
	}

	req := new(Request)
	err = json.Unmarshal(reqBody, req)
	if err != nil {
		log.Status(http.StatusBadRequest).Error("decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Ctx = r.Context()
	if req.Tx == "" {
		req.Tx = tx
	}
	log.Tx(req.Tx)

	u, err := url.Parse(req.URL)
	if err != nil {
		log.Status(http.StatusBadRequest).Error("decode request url: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.CSRF != "" {
		if csrf := GetCSRF(); csrf != "" {
			query := u.Query()
			query.Set(req.CSRF, csrf)
			u.RawQuery = query.Encode()
		}
	}
	err = wbiKeys.Sign(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Status(http.StatusInternalServerError).
			Cost(time.Since(start)).
			Error("wbi sign url: %v", err)
		return
	}
	req.URL = u.String()
	log.URL(req.URL)

	resp := DoRequestAPI(log, req)

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err = enc.Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Cost(time.Since(start)).Error("encode response: %v", err)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Cost(time.Since(start)).Error("write response: %v", err)
		return
	}
	log.Cost(time.Since(start)).Info("done")
}

func DoRequestAPI(log Logger, req *Request) (resp *Response) {
	tx := req.Tx
	ctx := req.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	resp = &Response{
		Tx:   tx,
		Body: json.RawMessage{'{', '}'},
	}

	if req.Method == "" {
		req.Method = "GET"
	}
	log.Method(req.Method).URL(req.URL)

	var body io.Reader = http.NoBody
	if json.Valid(req.Body) {
		body = bytes.NewReader(req.Body)
		log.Req(json.RawMessage(req.Body))
	}
	request, err := http.NewRequestWithContext(ctx, strings.ToUpper(req.Method), req.URL, body)
	if err != nil {
		log.Status(http.StatusInternalServerError).Error("new request: %v", err)
		resp.StatusCode = http.StatusInternalServerError
		return
	}
	request = HijackLogout(req.URL, request)
	setDefaultHeaders(request.Header)
	buvIdManager.AddCookies(request)
	headers := make(http.Header)
	for _, kvs := range req.Headers {
		headers.Add(kvs[0], kvs[1])
	}
	maps.Copy(request.Header, headers)
	AddLoginedCookies(request)
	log.ReqHeaders(request.Header)

	response, err := (&http.Client{
		Timeout: time.Duration(req.Timeout) * time.Millisecond,
	}).Do(request)
	if err != nil {
		if _, ok := err.(TimeoutError); ok {
			resp.StatusCode = http.StatusRequestTimeout
		} else {
			resp.StatusCode = http.StatusServiceUnavailable
		}
		log.Status(resp.StatusCode).Error("http request: %v", err)
		return
	}
	respBody, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		resp.StatusCode = http.StatusBadGateway
		log.Status(resp.StatusCode).Error("read response body: %v", err)
		return
	}
	if !json.Valid(respBody) {
		respBody, _ = json.Marshal(string(respBody))
	}
	log.Status(response.StatusCode).
		Rsp(json.RawMessage(respBody)).
		RspHeaders(response.Header)

	respHeaders := [][]string{}
	for key, values := range response.Header {
		for _, value := range values {
			respHeaders = append(respHeaders, []string{key, value})
		}
	}

	resp.StatusCode = response.StatusCode
	resp.Status = response.Status
	resp.Headers = respHeaders
	resp.Body = json.RawMessage(respBody)

	HijackPollAPI(req.URL, resp)
	HijackLogoutResp(req.URL, resp)
	HijackIsCookieExpire(resp)
	return
}
