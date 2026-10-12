package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/eachain/logfilewriter"
)

const LogsDir = "logs"

type LogRecord struct {
	Time       string         `json:"time"`
	Level      string         `json:"lv,omitempty"` // log level
	Tx         string         `json:"tx,omitempty"`
	File       string         `json:"file,omitempty"`
	Act        string         `json:"act,omitempty"`
	Extra      map[string]any `json:"extra,omitempty"`
	Method     string         `json:"method,omitempty"`
	URL        string         `json:"url,omitempty"`
	ReqHeaders map[string]any `json:"req_hdr,omitempty"`
	RspHeaders map[string]any `json:"rsp_hdr,omitempty"`
	Req        any            `json:"req,omitempty"`    // request pb/form/json ...
	Rsp        any            `json:"rsp,omitempty"`    // response pb/form/json ...
	Status     int            `json:"status,omitempty"` // http status code
	Cost       *int64         `json:"cost,omitempty"`   // unit: ms
	Msg        string         `json:"msg,omitempty"`    // log message
}

func RouteLogAPI(w http.ResponseWriter, r *http.Request) {
	lr := new(LogRecord)
	err := json.NewDecoder(r.Body).Decode(lr)
	if err != nil {
		return
	}

	lr.Time = time.Now().Format("2006-01-02 15:04:05.000")
	if lr.Level == "" {
		lr.Level = "info"
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(lr)
	if err != nil {
		return
	}
	globalLogWriter.Write(buf.Bytes())
}

var globalLogWriter io.Writer

func InitLogger(cacheDir string) {
	globalLogWriter = logfilewriter.New(
		logfilewriter.WithDir(filepath.Join(cacheDir, LogsDir)),
		logfilewriter.WithFileName("bilisten.log"),
		logfilewriter.WithFileSizeLimit(100*1024*1024), // 100MB
		logfilewriter.WithArchiveDir(filepath.Join(cacheDir, LogsDir, "archive")),
		logfilewriter.WithCompress(),
		logfilewriter.WithRotateDays(3),
	)
}

type Logger struct {
	r *LogRecord
}

func NewLogger(fn string) Logger {
	return Logger{r: &LogRecord{Act: fn}}
}

func (logger Logger) Tx(tx string) Logger {
	logger.r.Tx = tx
	return logger
}

func (logger Logger) Act(act string) Logger {
	if logger.r.Act == "" {
		logger.r.Act = act
	} else {
		logger.r.Act += ":" + act
	}
	return logger
}

func (logger Logger) Extra(key string, value any) Logger {
	if logger.r.Extra == nil {
		logger.r.Extra = make(map[string]any)
	}
	logger.r.Extra[key] = value
	return logger
}

func (logger Logger) Method(method string) Logger {
	logger.r.Method = method
	return logger
}

func (logger Logger) URL(url string) Logger {
	logger.r.URL = url
	return logger
}

func (logger Logger) ReqHeaders(headers http.Header) Logger {
	logger.r.ReqHeaders = make(map[string]any)
	for key, values := range headers {
		if len(values) == 1 {
			logger.r.ReqHeaders[key] = values[0]
		} else {
			logger.r.ReqHeaders[key] = values
		}
	}
	return logger
}

func (logger Logger) RspHeaders(headers http.Header) Logger {
	logger.r.RspHeaders = make(map[string]any)
	for key, values := range headers {
		if len(values) == 1 {
			logger.r.RspHeaders[key] = values[0]
		} else {
			logger.r.RspHeaders[key] = values
		}
	}
	return logger
}

func (logger Logger) Req(req any) Logger {
	logger.r.Req = req
	return logger
}

func (logger Logger) Rsp(rsp any) Logger {
	logger.r.Rsp = rsp
	return logger
}

func (logger Logger) Status(status int) Logger {
	logger.r.Status = status
	return logger
}

func (logger Logger) Cost(cost time.Duration) Logger {
	ms := cost.Milliseconds()
	logger.r.Cost = &ms
	return logger
}

func (logger Logger) Debug(format string, a ...any) {
	logger.output("debug", format, a...)
}

func (logger Logger) Info(format string, a ...any) {
	logger.output("info", format, a...)
}

func (logger Logger) Warn(format string, a ...any) {
	logger.output("warn", format, a...)
}

func (logger Logger) Error(format string, a ...any) {
	logger.output("error", format, a...)
}

func (logger Logger) afterOutput() {
	logger.r.Time = ""
	logger.r.File = ""
	logger.r.Level = ""
	logger.r.Msg = ""
	logger.r.Act = ""
	logger.r.Cost = nil
	logger.r.Status = 0
	logger.r.Req = nil
	logger.r.Rsp = nil
}

func (logger Logger) output(lv string, f string, a ...any) {
	defer logger.afterOutput()

	logger.r.Time = time.Now().Format("2006-01-02 15:04:05.000")

	_, file, line, ok := runtime.Caller(2)
	if ok {
		logger.r.File = fmt.Sprintf("%v:%v", filepath.Base(file), line)
	}

	logger.r.Level = lv
	logger.r.Msg = fmt.Sprintf(f, a...)

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(logger.r)
	if err != nil {
		return
	}
	globalLogWriter.Write(buf.Bytes())
}
