package dusk

import (
	"bytes"
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

	// EventRequest request event
	EventRequest = "request"
	// EventResponse response event
	EventResponse = "response"
	// EventError error event
	EventError = "error"
	// EventDone done event
	EventDone = "done"
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
		// EnableTimelineTrace enable the timeline trace
		EnableTimelineTrace bool
		// Request http request
		Request *http.Request
		// Response http response
		Response *http.Response
		// Error error
		Error error
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
	}
)

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

// Reset reset
func (d *Dusk) Reset() {
	d.Request = nil
	d.Response = nil
	d.Error = nil
	d.M = nil
	d.tl = nil
	d.events = nil
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

// Do do http request
func (d *Dusk) Do() (resp *http.Response, body []byte, err error) {
	defer func() {
		if err != nil {
			d.Emit(EventError)
		}
	}()
	defer d.Emit(EventDone)
	req := d.Request
	c := d.Client
	if c == nil {
		c = defaultClient
	}
	if d.EnableTimelineTrace {
		trace, tl := NewClientTrace()
		req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))
		d.Request = req
		d.tl = tl
	}
	d.Emit(EventRequest)
	// 如果在request event的处理函数中设置了error，出请求出错
	if d.Error != nil {
		err = d.Error
		return
	}
	resp, err = c.Do(req)
	if err != nil {
		d.Error = err
		return
	}
	d.Response = resp
	d.Emit(EventResponse)
	// 如果在response event的处理函数中设置了error，出请求出错
	if d.Error != nil {
		err = d.Error
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
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
	req.Header["Content-Type"] = []string{MIMEApplicationJSON}
}

// NewRequest new http request
func (d *Dusk) NewRequest(method, url string, query map[string]string, data interface{}, header map[string]string) (req *http.Request, err error) {
	// get new request url
	newURL := d.getURL(url, query)
	var r io.Reader
	// get send data reader
	if data != nil {
		buf, e := json.Marshal(data)
		if e != nil {
			err = e
			return
		}
		r = bytes.NewReader(buf)
	}
	req, err = http.NewRequest(method, newURL, r)
	if err != nil {
		return
	}
	// set content type
	if r != nil {
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

// On add event linster function
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

// New new a request
func New() *Dusk {
	// 是否需要直接在初始化时就生成，还是动态生成？
	return &Dusk{
		// M:                  make(map[string]interface{}),
	}
}
