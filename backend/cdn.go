package backend

import (
	"errors"
	"io"
	"log"
	"maps"
	"net/http"
	"net/http/httputil"
	"time"
)

var errInvalidWrite = errors.New("invalid write result")

type CDNOptions struct {
	Tx          string
	URL         string
	ContentType string
	Etag        string

	HookRespHeader func(http.Header)
	HookRespStatus func(int)
	HookRespBody   func([]byte)
	HookRespOver   func(error)
}

const CacheControl = "public, max-age=864000, immutable"

func ProxyCDN(w http.ResponseWriter, r *http.Request, opts CDNOptions) {
	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, opts.URL, nil)
	if err != nil {
		log.Printf("%v new request for url %q: %v", opts.Tx, opts.URL, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	setDefaultHeaders(request.Header)
	buvIdManager.AddCookies(request)
	for _, key := range []string{"accept", "accept-encoding", "accept-language", "range"} {
		if val := r.Header.Get(key); val != "" {
			request.Header.Set(key, val)
		}
	}

	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		log.Printf("%v dump request: %v", opts.Tx, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("%v dump request: %s", opts.Tx, dump)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("%v http request: %v", opts.Tx, err)
		if _, ok := err.(TimeoutError); ok {
			w.WriteHeader(http.StatusGatewayTimeout)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		return
	}
	defer response.Body.Close()

	maps.Copy(w.Header(), response.Header)
	AccessCors(w.Header())
	if w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", CacheControl)
	}
	if opts.ContentType != "" && w.Header().Get("Content-Type") == "application/octet-stream" {
		w.Header().Set("Content-Type", opts.ContentType)
	}
	if opts.Etag != "" && w.Header().Get("Etag") == "" {
		w.Header().Set("Etag", opts.Etag)
	}

	if opts.HookRespHeader != nil {
		opts.HookRespHeader(response.Header)
	}
	if opts.HookRespStatus != nil {
		opts.HookRespStatus(response.StatusCode)
	}
	if opts.HookRespOver != nil {
		defer func() { opts.HookRespOver(err) }()
	}

	w.WriteHeader(response.StatusCode)

	controller := http.NewResponseController(w)

	buf := make([]byte, 32*1024)
	var written int64 = 0
	for {
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			if opts.HookRespBody != nil {
				opts.HookRespBody(buf[0:nr])
			}
			nw, ew := w.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
			ew = controller.Flush()
			if ew != nil {
				err = ew
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	// read all
	for {
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			if opts.HookRespBody != nil {
				opts.HookRespBody(buf[0:nr])
			}
		}
		if er != nil {
			break
		}
	}

	if err != nil {
		log.Printf("%v response: %s, %v bytes, error: %v",
			opts.Tx, response.Status, written, err)
	} else {
		log.Printf("%v response: %s, %v bytes",
			opts.Tx, response.Status, written)
	}
}

func ReuseCache(w http.ResponseWriter, r *http.Request) (reuse bool) {
	if etag := r.Header.Get("if-none-match"); etag != "" {
		w.Header().Set("Etag", etag)
		reuse = true
	}
	if mtime := r.Header.Get("If-Modified-Since"); mtime != "" {
		w.Header().Set("Last-Modified", mtime)
		reuse = true
	}
	if reuse {
		w.Header().Set("Cache-Control", CacheControl)
		w.Header().Set("Date", time.Now().Format(http.TimeFormat))
		w.WriteHeader(http.StatusNotModified)
	}
	return
}
