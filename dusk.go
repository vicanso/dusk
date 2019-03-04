package dusk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	// 设置默认10秒超时
	defaultClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	// MIMEApplicationJSON application json
	MIMEApplicationJSON = "application/json"
	// HeaderContentType content type
	HeaderContentType = "Content-Type"

	// EventRequest request event
	EventRequest = "request"
	// EventResponse response event
	EventResponse = "response"
	// EventError error event
	EventError = "error"
	// EventDone done event
	EventDone = "done"

	gzipEncoding    = "gzip"
	contentEncoding = "Content-Encoding"
)

type (
	// Listener event listener
	Listener func(*Dusk)
	// Event event struct
	Event struct {
		Name     string
		Listener Listener
	}
	// Dusk http request
	Dusk struct {
		// Timeout timeout for request
		Timeout time.Duration
		// EnableTimelineTrace enable the timeline trace
		EnableTimelineTrace bool
		// Request http request
		Request *http.Request
		// Response http response
		Response *http.Response
		// Body response body
		Body []byte
		// RequestBody request body
		RequestBody []byte
		// Error error
		Error error
		// ConvertError convert error
		ConvertError func(error, *Dusk) error
		// Client http client
		Client *http.Client
		// URLPrefix the prefix of request url
		URLPrefix string
		// M data for dusk
		M map[string]interface{}
		// http timeline struct
		tl *HTTPTimeline
		// events list
		events []*Event
		ctx    context.Context
	}
)

// Reset reset
func (d *Dusk) Reset() {
	d.Timeout = 0
	d.EnableTimelineTrace = false
	d.Request = nil
	d.Response = nil
	d.Body = nil
	d.RequestBody = nil
	d.Error = nil
	d.ConvertError = nil
	d.Client = nil
	d.URLPrefix = ""
	d.M = nil
	d.tl = nil
	d.events = nil
	d.ctx = nil
}

// GetURL get the url with query string
func GetURL(u string, query map[string]string) string {
	if query == nil {
		return u
	}
	p := url.Values{}
	for k, v := range query {
		p.Set(k, v)
	}
	if strings.Contains(u, "?") {
		return u + "&" + p.Encode()
	}
	return u + "?" + p.Encode()
}

// SetValue set value
func (d *Dusk) SetValue(k string, v interface{}) {
	if d.M == nil {
		d.M = make(map[string]interface{})
	}
	d.M[k] = v
}

// GetValue get value
func (d *Dusk) GetValue(k string) interface{} {
	return d.M[k]
}

func (d *Dusk) do() {
	d.Emit(EventRequest)
	// 此处的Request有可能会在 request event中被调整
	req := d.Request
	c := d.Client
	if c == nil {
		c = defaultClient
	}
	if d.EnableTimelineTrace {
		trace, tl := NewClientTrace()
		ctx := d.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		d.ctx = httptrace.WithClientTrace(ctx, trace)
		req = req.WithContext(d.ctx)
		d.Request = req
		d.tl = tl
	}

	// 如果在request event 的处理函数中设置了error，出请求出错
	if d.Error != nil {
		return
	}
	var resp *http.Response
	resp, d.Error = c.Do(req)
	if d.Error != nil {
		return
	}
	d.Response = resp
	defer resp.Body.Close()
	var reader io.ReadCloser
	switch resp.Header.Get(contentEncoding) {
	case gzipEncoding:
		reader, d.Error = gzip.NewReader(resp.Body)
		if d.Error != nil {
			return
		}
		resp.Header.Del(contentEncoding)
	default:
		reader = resp.Body
	}

	var buf []byte
	buf, d.Error = ioutil.ReadAll(reader)
	if d.Error != nil {
		return
	}

	d.Body = buf
	d.Emit(EventResponse)
	return
}

// Do do http request
func (d *Dusk) Do() (resp *http.Response, body []byte, err error) {
	d.do()
	resp = d.Response
	body = d.Body
	if d.Error != nil {
		if d.ConvertError != nil {
			e := d.ConvertError(d.Error, d)
			if e != nil {
				d.Error = e
			}
		}
		d.Emit(EventError)
		err = d.Error
	}
	d.Emit(EventDone)
	return
}

func (d *Dusk) fillHeader(req *http.Request, header map[string]string) {
	currentHeader := req.Header
	for k, v := range header {
		if len(currentHeader[k]) == 0 {
			currentHeader[k] = make([]string, 0)
		}
		currentHeader[k] = append(currentHeader[k], v)
	}
}

// get request url
func (d *Dusk) getURL(url string, query map[string]string) string {
	newURL := GetURL(url, query)
	if d.URLPrefix != "" {
		newURL = d.URLPrefix + newURL
	}
	return newURL
}

// setJsonContentType set json content type
func (d *Dusk) setJSONContentType(req *http.Request) {
	req.Header[HeaderContentType] = []string{MIMEApplicationJSON}
}

// NewRequest new http request
func (d *Dusk) NewRequest(method, url string, query map[string]string, data interface{}, header map[string]string) (req *http.Request, err error) {
	// get new request url
	newURL := d.getURL(url, query)
	var r io.Reader
	isJSON := false
	// get send data reader
	if data != nil {
		v, ok := data.(io.Reader)
		if ok {
			r = v
		} else {
			// 如果非reader 序列化为json
			buf, e := json.Marshal(data)
			if e != nil {
				err = e
				return
			}
			d.RequestBody = buf
			r = bytes.NewReader(buf)
			isJSON = true
		}
	}
	req, err = http.NewRequest(method, newURL, r)

	// 如果有设置超时，则调整context
	if d.Timeout != 0 {
		currentCtx := d.ctx
		if currentCtx == nil {
			currentCtx = context.Background()
		}
		ctx, cancel := context.WithTimeout(currentCtx, d.Timeout)
		d.ctx = ctx
		d.On(EventDone, func(_ *Dusk) {
			cancel()
		})
	}
	if d.ctx != nil {
		req = req.WithContext(d.ctx)
	}
	if err != nil {
		return
	}
	// set json content type
	if isJSON {
		d.setJSONContentType(req)
	}
	d.Request = req

	if header != nil {
		d.fillHeader(req, header)
	}
	return
}

// Get http get request
func (d *Dusk) Get(url string, query map[string]string) (resp *http.Response, body []byte, err error) {
	return d.GetWithHeader(url, query, nil)
}

// GetWithHeader http get request with headers
func (d *Dusk) GetWithHeader(url string, query, header map[string]string) (resp *http.Response, body []byte, err error) {
	_, err = d.NewRequest(http.MethodGet, url, query, nil, header)
	if err != nil {
		return
	}
	return d.Do()
}

// Post the http post request
func (d *Dusk) Post(url string, data interface{}, query map[string]string) (resp *http.Response, body []byte, err error) {
	return d.PostWithHeader(url, data, query, nil)
}

// PostWithHeader post request with header
func (d *Dusk) PostWithHeader(url string, data interface{}, query, header map[string]string) (resp *http.Response, body []byte, err error) {
	_, err = d.NewRequest(http.MethodPost, url, query, data, header)
	if err != nil {
		return
	}
	return d.Do()
}

// Patch http patch request
func (d *Dusk) Patch(url string, data interface{}, query map[string]string) (resp *http.Response, body []byte, err error) {
	return d.PatchWithHeader(url, data, query, nil)
}

// PatchWithHeader patch with header
func (d *Dusk) PatchWithHeader(url string, data interface{}, query, header map[string]string) (resp *http.Response, body []byte, err error) {
	_, err = d.NewRequest(http.MethodPatch, url, query, data, header)
	if err != nil {
		return
	}
	return d.Do()
}

// Put http put request
func (d *Dusk) Put(url string, data interface{}, query map[string]string) (resp *http.Response, body []byte, err error) {
	return d.PutWithHeader(url, data, query, nil)
}

// PutWithHeader put with header
func (d *Dusk) PutWithHeader(url string, data interface{}, query, header map[string]string) (resp *http.Response, body []byte, err error) {
	_, err = d.NewRequest(http.MethodPut, url, query, data, header)
	if err != nil {
		return
	}
	return d.Do()
}

// Del del request
func (d *Dusk) Del(url string, query map[string]string) (resp *http.Response, body []byte, err error) {
	return d.DelWithHeader(url, query, nil)
}

// DelWithHeader delrequest with header
func (d *Dusk) DelWithHeader(url string, query, header map[string]string) (resp *http.Response, body []byte, err error) {
	_, err = d.NewRequest(http.MethodDelete, url, query, nil, header)
	if err != nil {
		return
	}
	return d.Do()
}

// GetTimelineStats get the timeline stats
func (d *Dusk) GetTimelineStats() *HTTPTimelineStats {
	if d.tl == nil {
		return nil
	}
	return d.tl.Stats()
}

// On add event listen function
func (d *Dusk) On(name string, ln Listener) {
	if d.events == nil {
		d.events = make([]*Event, 0)
	}
	d.events = append(d.events, &Event{
		Name:     name,
		Listener: ln,
	})
}

// Emit emit event
func (d *Dusk) Emit(name string) {
	for _, e := range d.events {
		if e.Name == name {
			e.Listener(d)
		}
	}
}

// SetContext set context to dusk
func (d *Dusk) SetContext(ctx context.Context) {
	d.ctx = ctx
}

// GetContext get context of dusk
func (d *Dusk) GetContext() context.Context {
	return d.ctx
}

// New new a request
func New() *Dusk {
	// 是否需要直接在初始化时就生成，还是动态生成？
	return &Dusk{
		// M:                  make(map[string]interface{}),
	}
}
