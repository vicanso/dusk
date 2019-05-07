// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dusk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/dsnet/compress/brotli"
	"github.com/golang/snappy"
)

const (
	// MIMEApplicationJSON application json
	MIMEApplicationJSON = "application/json"
	// MIMEApplicationFormUrlencoded form url encoded
	MIMEApplicationFormUrlencoded = "application/x-www-form-urlencoded"
	// HeaderContentType content type
	HeaderContentType = "Content-Type"
	// HeaderContentEncoding content encoding
	HeaderContentEncoding = "Content-Encoding"
	// HeaderContentLength content length
	HeaderContentLength = "Content-Length"
	// HeaderAcceptEncoding accept encoding
	HeaderAcceptEncoding = "Accept-Encoding"
	// GzipEncoding gzip encoding
	GzipEncoding = "gzip"
	// SnappyEncoding snappy encoding
	SnappyEncoding = "snappy"
	// BrEncoding br encoding
	BrEncoding = "br"

	jsonType = "json"
	formType = "form"

	httpProtocol  = "http://"
	httpsProtocol = "https://"
)

const (
	// EventTypeNone none event type
	EventTypeNone = iota
	// EventTypeBefore before event
	EventTypeBefore
	// EventTypeAfter after event
	EventTypeAfter
)

var (
	globalRequestEvents  []*RequestEvent
	globalResponseEvents []*ResponseEvent
	globalErrorListeners []ErrorListener
	doneListeners        []DoneListener

	// defaultConfig default config for all request
	defaultConfig *Config
)

type (
	// Config the config for request
	Config struct {
		// BaseURL it will be prepended to url unless url is absolute.
		BaseURL string
		// Headers it will be added to request header
		Headers http.Header
		// Timeout timeout for request
		Timeout time.Duration
	}
	// Decoder compression decoder
	Decoder func(*http.Response) ([]byte, error)
	// DoneListener done event listener
	DoneListener func(*Dusk) error
	// RequestListener request event listener
	RequestListener func(*http.Request, *Dusk) (newErr error)
	// ResponseListener response event listener
	ResponseListener func(*http.Response, *Dusk) (newErr error)
	// ErrorListener error event listener
	ErrorListener func(error, *Dusk) (newErr error)

	// Dusk http request client
	Dusk struct {
		// Request http request
		Request *http.Request
		// Response http response
		Response *http.Response
		// Body response's body
		Body []byte
		// Err request error
		Err error

		client         *http.Client
		m              map[string]interface{}
		header         http.Header
		params         map[string]string
		query          url.Values
		data           interface{}
		ctx            context.Context
		doneListeners  []DoneListener
		requestEvents  []*RequestEvent
		responseEvents []*ResponseEvent
		errorListeners []ErrorListener
		url            string
		path           string
		method         string
		timeout        time.Duration
		ht             *HTTPTrace
		enabledTrace   bool
	}
	// RequestEvent request event
	RequestEvent struct {
		ln RequestListener
		t  int
	}
	// ResponseEvent response event
	ResponseEvent struct {
		ln ResponseListener
		t  int
	}
)

// AddRequestListener add request listener for all http requset,
// it will be called before or after http request.
// If return new request, it will be overrded the original request.
// If return new error, it will return error and abort request.
func AddRequestListener(ln RequestListener, eventType int) {
	if globalRequestEvents == nil {
		globalRequestEvents = make([]*RequestEvent, 0)
	}
	globalRequestEvents = append(globalRequestEvents, &RequestEvent{
		ln: ln,
		t:  eventType,
	})
}

// ClearRequestListener clear global request listener
func ClearRequestListener() {
	globalRequestEvents = nil
}

// AddResponseListener add response listener for all http requset,
// it will be called before or after http response.
// If return new response, it will be overried the original response.
// If return new error, it will return error and abort response.
func AddResponseListener(ln ResponseListener, eventType int) {
	if globalResponseEvents == nil {
		globalResponseEvents = make([]*ResponseEvent, 0)
	}
	globalResponseEvents = append(globalResponseEvents, &ResponseEvent{
		ln: ln,
		t:  eventType,
	})
}

// ClearResponseListener clear response listener
func ClearResponseListener() {
	globalResponseEvents = nil
}

// AddErrorListener add error listener for all http request
func AddErrorListener(ln ErrorListener) {
	if globalErrorListeners == nil {
		globalErrorListeners = make([]ErrorListener, 0)
	}
	globalErrorListeners = append(globalErrorListeners, ln)
}

// ClearErrorListener clear all http error listener
func ClearErrorListener() {
	globalErrorListeners = nil
}

// AddDoneListener add done listener
func AddDoneListener(lnList ...DoneListener) {
	if doneListeners == nil {
		doneListeners = make([]DoneListener, 0)
	}
	doneListeners = append(doneListeners, lnList...)
}

func getClient(d *Dusk) *http.Client {
	c := d.client
	if c == nil {
		c = http.DefaultClient
	}
	return c
}

func snappyDecoder(resp *http.Response) (buf []byte, err error) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var dst []byte
	buf, err = snappy.Decode(dst, data)
	if err != nil {
		return
	}
	return
}

// SnappyDecode support snappy decode for response,
// if the Content-Encoding:snappy, the docode function will be called
func SnappyDecode(resp *http.Response, d *Dusk) (newErr error) {
	return decode(resp, d, SnappyEncoding, snappyDecoder)
}

func decode(resp *http.Response, d *Dusk, encoding string, decoder Decoder) (newErr error) {
	if resp.Header.Get(HeaderContentEncoding) != encoding {
		return
	}

	resp.Uncompressed = true
	resp.Header.Del(HeaderContentEncoding)
	resp.Header.Del(HeaderContentLength)

	buf, err := decoder(resp)
	if err != nil {
		newErr = err
		return
	}
	d.Body = buf
	return
}

func brDecoder(resp *http.Response) (buf []byte, err error) {
	defer resp.Body.Close()
	r, err := brotli.NewReader(resp.Body, new(brotli.ReaderConfig))
	if err != nil {
		return
	}
	buf, err = ioutil.ReadAll(r)
	return
}

// BrDecode support brotli decode for response,
// if the Content-Encoding:br, the docode function will be called
func BrDecode(resp *http.Response, d *Dusk) (newErr error) {
	return decode(resp, d, BrEncoding, brDecoder)
}

// SetClient set http client for dusk
func (d *Dusk) SetClient(client *http.Client) *Dusk {
	d.client = client
	return d
}

// GetClient get http client of dusk
func (d *Dusk) GetClient() *http.Client {
	return d.client
}

// SetValue set value
func (d *Dusk) SetValue(k string, v interface{}) *Dusk {
	if d.m == nil {
		d.m = make(map[string]interface{})
	}
	d.m[k] = v
	return d
}

// GetValue get value
func (d *Dusk) GetValue(k string) interface{} {
	return d.m[k]
}

// Set set http request header
func (d *Dusk) Set(key, value string) *Dusk {
	if d.header == nil {
		d.header = make(http.Header)
	}
	d.header.Set(key, value)
	return d
}

// Type set the content type of request
func (d *Dusk) Type(contentType string) *Dusk {
	switch contentType {
	case jsonType:
		contentType = MIMEApplicationJSON
	case formType:
		contentType = MIMEApplicationFormUrlencoded
	}
	d.Set(HeaderContentType, contentType)
	return d
}

// Queries set http request query
func (d *Dusk) Queries(query map[string]string) *Dusk {
	for k, v := range query {
		d.Query(k, v)
	}
	return d
}

// Query set http request query
func (d *Dusk) Query(key, value string) *Dusk {
	if d.query == nil {
		d.query = make(url.Values)
	}
	d.query.Set(key, value)
	return d
}

// Param set http request url param
func (d *Dusk) Param(key, value string) *Dusk {
	if d.params == nil {
		d.params = make(map[string]string)
	}
	d.params[key] = value
	return d
}

// Send set the send data
func (d *Dusk) Send(data interface{}) *Dusk {
	d.data = data
	return d
}

// SetContext set context to dusk
func (d *Dusk) SetContext(ctx context.Context) *Dusk {
	d.ctx = ctx
	return d
}

// GetContext get context of dusk
func (d *Dusk) GetContext() context.Context {
	return d.ctx
}

// Timeout set timeout for request
func (d *Dusk) Timeout(timeout time.Duration) *Dusk {
	d.timeout = timeout
	return d
}

// AddDoneListener add done listener
func (d *Dusk) AddDoneListener(lnList ...DoneListener) *Dusk {
	if d.doneListeners == nil {
		d.doneListeners = make([]DoneListener, 0)
	}
	d.doneListeners = append(d.doneListeners, lnList...)
	return d
}

// EmitDone emit done event
func (d *Dusk) EmitDone() error {
	size := len(d.doneListeners)
	if size == 0 {
		return nil
	}
	for i := size - 1; i >= 0; i-- {
		ln := d.doneListeners[i]
		err := ln(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dusk) addRequestEvent(events ...*RequestEvent) *Dusk {
	if d.requestEvents == nil {
		d.requestEvents = make([]*RequestEvent, 0)
	}
	d.requestEvents = append(d.requestEvents, events...)
	return d
}

// AddRequestListener add request listene
func (d *Dusk) AddRequestListener(ln RequestListener, eventType int) *Dusk {
	return d.addRequestEvent(&RequestEvent{
		ln: ln,
		t:  eventType,
	})
}

// EmitRequest emit request event
func (d *Dusk) EmitRequest(t int) error {
	size := len(d.requestEvents)
	if size == 0 {
		return nil
	}
	// 从后往前执行，后加入的先执行
	// 本请求的 --> instance --> global
	for i := size - 1; i >= 0; i-- {
		e := d.requestEvents[i]
		if e.t != t {
			continue
		}
		err := e.ln(d.Request, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dusk) addResponseEvent(events ...*ResponseEvent) *Dusk {
	if d.responseEvents == nil {
		d.responseEvents = make([]*ResponseEvent, 0)
	}
	d.responseEvents = append(d.responseEvents, events...)
	return d
}

// AddResponseListener add response listener
func (d *Dusk) AddResponseListener(ln ResponseListener, eventType int) *Dusk {
	return d.addResponseEvent(&ResponseEvent{
		ln: ln,
		t:  eventType,
	})
}

// EmitResponse emit response event
func (d *Dusk) EmitResponse(t int) error {
	size := len(d.responseEvents)
	if size == 0 {
		return nil
	}
	for i := size - 1; i >= 0; i-- {
		e := d.responseEvents[i]
		if e.t != t {
			continue
		}
		err := e.ln(d.Response, d)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddErrorListener add error listener
func (d *Dusk) AddErrorListener(lnList ...ErrorListener) *Dusk {
	if d.errorListeners == nil {
		d.errorListeners = make([]ErrorListener, 0)
	}
	d.errorListeners = append(d.errorListeners, lnList...)
	return d
}

// EmitError emit error event
func (d *Dusk) EmitError(currentErr error) error {
	for _, ln := range d.errorListeners {
		err := ln(currentErr, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func prependURL(requestURL string, config *Config) string {
	// 如果有配置了base url，而且当前请求不是以绝对路径
	if config != nil && config.BaseURL != "" {
		if !(strings.HasPrefix(requestURL, httpProtocol) || strings.HasPrefix(requestURL, httpsProtocol)) {
			requestURL = config.BaseURL + requestURL
		}
	}
	return requestURL
}

func newDusk(method, requestURL string) *Dusk {
	requestURL = prependURL(requestURL, defaultConfig)

	info, _ := url.Parse(requestURL)
	path := ""
	if info != nil {
		path = info.Path
	}
	d := &Dusk{
		url:    requestURL,
		path:   path,
		method: method,
	}
	if defaultConfig != nil && defaultConfig.Timeout != 0 {
		d.Timeout(defaultConfig.Timeout)
	}

	if globalRequestEvents != nil {
		d.addRequestEvent(globalRequestEvents...)
	}
	if globalResponseEvents != nil {
		d.addResponseEvent(globalResponseEvents...)
	}
	if globalErrorListeners != nil {
		d.AddErrorListener(globalErrorListeners...)
	}
	if doneListeners != nil {
		d.AddDoneListener(doneListeners...)
	}

	return d
}

// Get http get request
func Get(url string) *Dusk {
	return newDusk(http.MethodGet, url)
}

// Head http head request
func Head(url string) *Dusk {
	return newDusk(http.MethodHead, url)
}

// Post http post request
func Post(url string) *Dusk {
	return newDusk(http.MethodPost, url)
}

// Put http put request
func Put(url string) *Dusk {
	return newDusk(http.MethodPut, url)
}

// Patch http patch request
func Patch(url string) *Dusk {
	return newDusk(http.MethodPatch, url)
}

// Delete http delete request
func Delete(url string) *Dusk {
	return newDusk(http.MethodDelete, url)
}

// 添加 config 中配置的http头
func addConfigHeader(req *http.Request, config *Config) {
	if config != nil {
		for key, values := range config.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
}

func (d *Dusk) newRequest() (req *http.Request, err error) {
	data := d.data
	var r io.Reader
	// get send data reader
	if data != nil {
		v, ok := data.(io.Reader)
		if ok {
			r = v
		} else {
			values, ok := data.(url.Values)
			// 如果是form，则序列化为 x-www-form-urlencoded
			if ok {
				d.Type(formType)
				r = bytes.NewReader([]byte(values.Encode()))
			} else {
				// 如果非reader 序列化为json
				buf, e := json.Marshal(data)
				if e != nil {
					err = e
					return
				}
				r = bytes.NewReader(buf)
			}
		}
		// 如果没有设置 content-type 默认为 json
		if d.header == nil || d.header.Get(HeaderContentType) == "" {
			d.Type(jsonType)
		}
	}
	req, err = http.NewRequest(d.method, d.GetURL(), r)
	if err != nil {
		return
	}
	addConfigHeader(req, defaultConfig)
	// 如果有设置超时，则调整context
	if d.timeout != 0 {
		currentCtx := d.ctx
		if currentCtx == nil {
			currentCtx = context.Background()
		}
		ctx, cancel := context.WithTimeout(currentCtx, d.timeout)
		d.ctx = ctx
		d.AddDoneListener(func(_ *Dusk) error {
			cancel()
			return nil
		})
	}
	if d.ctx != nil {
		req = req.WithContext(d.ctx)
	}
	if err != nil {
		return
	}
	for k, values := range d.header {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}
	return
}

// EnableTrace enable trace
func (d *Dusk) EnableTrace() *Dusk {
	d.enabledTrace = true
	return d
}

// GetHTTPTrace get http trace
func (d *Dusk) GetHTTPTrace() *HTTPTrace {
	return d.ht
}

func (d *Dusk) addAcceptEncoding(encoding string) {
	accept := ""
	header := d.header
	if header != nil {
		accept = header.Get(HeaderAcceptEncoding)
	}
	// gzip is support by default
	if accept == "" {
		accept = GzipEncoding
	}
	accept += (", " + encoding)
	d.Set(HeaderAcceptEncoding, accept)
	return
}

// Snappy add snappy decode response
func (d *Dusk) Snappy() *Dusk {
	if d.isDisableCompression() {
		return d
	}
	d.addAcceptEncoding(SnappyEncoding)
	d.AddResponseListener(SnappyDecode, EventTypeBefore)
	return d
}

// Br add brotli decode response
func (d *Dusk) Br() *Dusk {
	if d.isDisableCompression() {
		return d
	}
	d.addAcceptEncoding(BrEncoding)
	d.AddResponseListener(BrDecode, EventTypeBefore)
	return d
}

func (d *Dusk) isDisableCompression() bool {
	c := getClient(d)
	if c.Transport != nil {
		if t, ok := c.Transport.(*http.Transport); ok {
			if t.DisableCompression {
				return true
			}
		}
	}
	return false
}

func (d *Dusk) do() (err error) {
	req := d.Request
	c := getClient(d)
	err = d.EmitRequest(EventTypeBefore)
	// 如果启用trace ，则添加相应的 context
	if d.enabledTrace {
		trace, ht := NewClientTrace()
		defer ht.Finish()
		ctx := d.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		d.ctx = httptrace.WithClientTrace(ctx, trace)
		req = req.WithContext(d.ctx)
		d.Request = req
		d.ht = ht
	}
	if err != nil {
		return
	}
	resp, err := c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = d.EmitRequest(EventTypeAfter)
	if err != nil {
		return
	}
	d.Response = resp
	// 触发 response 事件
	err = d.EmitResponse(EventTypeBefore)
	if err != nil {
		return
	}
	// 因此在 response 事件中有可能会生成新的 response
	resp = d.Response
	// 如果已获取到数据，则返回
	if d.Body != nil {
		return
	}

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	d.Body = buf
	// 触发 response 事件
	err = d.EmitResponse(EventTypeAfter)
	if err != nil {
		return
	}

	return
}

// Do do http request
func (d *Dusk) Do() (resp *http.Response, body []byte, err error) {
	done := func() {
		if err != nil {
			newErr := d.EmitError(err)
			if newErr != nil {
				err = newErr
			}
		}
		e := d.EmitDone()
		if e != nil {
			err = e
		}
		d.Err = err
	}

	req, err := d.newRequest()
	if err != nil {
		done()
		return
	}
	d.Request = req
	err = d.do()
	if err != nil {
		done()
		return
	}
	resp = d.Response
	body = d.Body
	done()
	return
}

// GetMethod get request method
func (d *Dusk) GetMethod() string {
	return d.method
}

// GetURL get request url
func (d *Dusk) GetURL() string {
	url := d.url
	for key, value := range d.params {
		url = strings.Replace(url, ":"+key, value, -1)
	}
	if d.query != nil {
		qs := d.query.Encode()
		if strings.Contains(url, "?") {
			url += ("&" + qs)
		} else {
			url += ("?" + qs)
		}
	}
	return url
}

// GetPath get path of request
func (d *Dusk) GetPath() string {
	return d.path
}

// SetConfig set config
func SetConfig(c Config) {
	defaultConfig = &c
}
