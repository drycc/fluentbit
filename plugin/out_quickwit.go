// drycc quickwit plugin
package main

import (
	"C"
)

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
	"unsafe"

	"github.com/Masterminds/sprig/v3"
	"github.com/fluent/fluent-bit-go/output"
)

var (
	MaxLen    int64
	Revision  string
	BuildDate string

	// plugin config
	BaseURL        string
	IndexName      string
	Compress       bool
	BufferSize     int
	JSONDateKey    string
	JSONDateFormat string
	// internal variables
	BufferPool        sync.Pool
	IndexNameTemplate *template.Template
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	fmt.Printf("Quickwit output version %s %s", Revision, BuildDate)
	return output.FLBPluginRegister(ctx, "quickwit", "Ship fluent-bit logs to quickwit")
}

// (fluentbit will call this)
// ctx (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	BaseURL = output.FLBPluginConfigKey(ctx, "Base_URL")
	IndexName = output.FLBPluginConfigKey(ctx, "Index_Name")
	Compress, _ = strconv.ParseBool(output.FLBPluginConfigKey(ctx, "Compress"))

	var err error
	BufferSize, err = strconv.Atoi(output.FLBPluginConfigKey(ctx, "Buffer_Size"))
	if err != nil {
		panic(err)
	}
	BufferPool = sync.Pool{
		New: func() any {
			buffer := &bytes.Buffer{}
			buffer.Grow(BufferSize)
			return buffer
		},
	}
	IndexNameTemplate, err = template.New("index").Funcs(sprig.FuncMap()).Parse(IndexName)
	if err != nil {
		panic(err)
	}
	JSONDateKey = output.FLBPluginConfigKey(ctx, "Json_Date_Key")
	JSONDateFormat = output.FLBPluginConfigKey(ctx, "Json_Date_Format")
	return output.FLB_OK
}

//export FLBPluginFlushCtx
//revive:disable:unused-parameter
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, _ *C.char) int {
	decoder := output.NewDecoder(data, int(length))
	results := make(map[string]*bytes.Buffer)
	for {
		ret, ts, rec := output.GetRecord(decoder)
		if ret != 0 {
			break
		}
		data := convertMap(rec)
		data[JSONDateKey] = formatTime(ts, JSONDateFormat)
		if jsonBytes, err := json.Marshal(data); err == nil {
			var buffer bytes.Buffer
			IndexNameTemplate.Execute(&buffer, data)
			index := buffer.String()
			if buffer, ok := results[index]; !ok {
				buffer := BufferPool.Get().(*bytes.Buffer)
				results[index] = buffer
				buffer.Write(jsonBytes)
				buffer.WriteByte('\n')
			} else {
				buffer.Write(jsonBytes)
				buffer.WriteByte('\n')
			}
		}
	}
	if err := sendResults(results); err != nil {
		return output.FLB_ERROR
	}
	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func formatTime(ts any, format string) string {
	var timestamp time.Time
	switch t := ts.(type) {
	case output.FLBTime:
		timestamp = ts.(output.FLBTime).Time
	case uint64:
		timestamp = time.Unix(int64(t), 0)
	default:
		timestamp = time.Now()
	}
	switch format {
	case "rfc3399":
		return timestamp.Format(time.RFC3339)
	case "unix_timestamp":
		return fmt.Sprint(timestamp.Unix())
	default:
		return timestamp.Format(strings.NewReplacer(
			"%Y", "2006",
			"%y", "06",
			"%m", "01",
			"%d", "02",
			"%H", "15",
			"%I", "03",
			"%M", "04",
			"%S", "05",
			"%p", "PM",
			"%L", "000",
			"%f", "000000",
			"%z", "-0700",
			"%Z", "MST",
		).Replace(format))
	}
}

func convertMap(values map[any]any) map[string]any {
	m := make(map[string]any)
	for k, v := range values {
		key := fmt.Sprintf("%v", k)
		switch t := v.(type) {
		case []byte:
			m[key] = string(t)
		case map[any]any:
			m[key] = convertMap(t)
		default:
			m[key] = v
		}
	}
	return m
}

func sendResults(results map[string]*bytes.Buffer) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: false},
			ForceAttemptHTTP2:  true,
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: !Compress,
		},
		Timeout: 30 * time.Second,
	}
	defer client.CloseIdleConnections()
	for index, body := range results {
		url := fmt.Sprintf("%s/api/v1/%s/ingest", BaseURL, index)
		if req, err := http.NewRequest("POST", url, body); err == nil {
			if resp, err := client.Do(req); err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				body.Reset()
			}
		}
	}
	return nil
}

func main() {}
