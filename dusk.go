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

	jsonType = "json"
	formType = "form"
)

const (
	// EventTypeNone none event type
	EventTypeNone = iota
	// EventTypeBefore before event
	EventTypeBefore
	// EventTypeAfter after event
	EventTypeAfter
)

type (
	// DoneListener done event listener
	DoneListener func(*Dusk) error
	// RequestListener request event listener
	RequestListener func(*http.Request, *Dusk) (newReq *http.Request, newErr error)
	// ResponseListener response event listener
	ResponseListener func(*http.Response, *Dusk) (newResp *http.Response, newErr error)
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
		query          url.Values
		data           interface{}
		ctx            context.Context
		doneEvents     []DoneListener
		requestEvents  []*RequestEvent
		responseEvents []*ResponseEvent
		errorEvents    []ErrorListener
		url            string
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

// SnappyDecode decode snappy response
func SnappyDecode(resp *http.Response, d *Dusk) (newResp *http.Response, newErr error) {
	if resp.Header.Get(HeaderContentEncoding) != SnappyEncoding {
		return
	}
	resp.Header.Del(HeaderContentEncoding)
	resp.Header.Del(HeaderContentLength)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		newErr = err
		return
	}
	var dst []byte
	buf, err := snappy.Decode(dst, data)
	if err != nil {
		newErr = err
		return
	}
	d.Body = buf
	return
}

// SetClient set client for dusk
func (d *Dusk) SetClient(client *http.Client) *Dusk {
	d.client = client
	return d
}

// GetClient get client of dusk
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

// OnDone on done event
func (d *Dusk) OnDone(ln DoneListener) *Dusk {
	if d.doneEvents == nil {
		d.doneEvents = make([]DoneListener, 0)
	}
	d.doneEvents = append(d.doneEvents, ln)
	return d
}

// EmitDone emit done event
func (d *Dusk) EmitDone() error {
	for _, ln := range d.doneEvents {
		err := ln(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dusk) addRequestListener(ln RequestListener, t int) *Dusk {
	if d.requestEvents == nil {
		d.requestEvents = make([]*RequestEvent, 0)
	}
	d.requestEvents = append(d.requestEvents, &RequestEvent{
		t:  t,
		ln: ln,
	})
	return d
}

// OnRequest on request event
func (d *Dusk) OnRequest(ln RequestListener) *Dusk {
	return d.addRequestListener(ln, EventTypeBefore)
}

// OnRequestSuccess on request success event
func (d *Dusk) OnRequestSuccess(ln RequestListener) *Dusk {
	return d.addRequestListener(ln, EventTypeAfter)
}

// EmitRequest emit request event
func (d *Dusk) EmitRequest(t int) error {
	for _, e := range d.requestEvents {
		if e.t != t {
			continue
		}
		newReq, err := e.ln(d.Request, d)
		if err != nil {
			return err
		}
		if newReq != nil {
			d.Request = newReq
		}
	}
	return nil
}

func (d *Dusk) addResponseListener(ln ResponseListener, t int) *Dusk {
	if d.responseEvents == nil {
		d.responseEvents = make([]*ResponseEvent, 0)
	}
	d.responseEvents = append(d.responseEvents, &ResponseEvent{
		t:  t,
		ln: ln,
	})
	return d
}

// OnResponse on response event
func (d *Dusk) OnResponse(ln ResponseListener) *Dusk {
	return d.addResponseListener(ln, EventTypeBefore)
}

// OnResponseSuccess on response success event
func (d *Dusk) OnResponseSuccess(ln ResponseListener) *Dusk {
	return d.addResponseListener(ln, EventTypeAfter)
}

// EmitResponse emit response event
func (d *Dusk) EmitResponse(t int) error {
	for _, e := range d.responseEvents {
		if e.t != t {
			continue
		}
		newResp, err := e.ln(d.Response, d)
		if err != nil {
			return err
		}
		if newResp != nil {
			d.Response = newResp
		}
	}
	return nil
}

// OnError on error event
func (d *Dusk) OnError(ln ErrorListener) *Dusk {
	if d.errorEvents == nil {
		d.errorEvents = make([]ErrorListener, 0)
	}
	d.errorEvents = append(d.errorEvents, ln)
	return d
}

// EmitError emit error event
func (d *Dusk) EmitError(currentErr error) error {
	for _, ln := range d.errorEvents {
		err := ln(currentErr, d)
		if err != nil {
			return err
		}
	}
	return nil
}

func newDusk(method, url string) *Dusk {
	return &Dusk{
		url:    url,
		method: method,
	}
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

func (d *Dusk) newReqest() (req *http.Request, err error) {
	url := d.url
	if d.query != nil {
		qs := d.query.Encode()
		if strings.Contains(url, "?") {
			url += ("&" + qs)
		} else {
			url += ("?" + qs)
		}
	}
	data := d.data
	var r io.Reader
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
			r = bytes.NewReader(buf)
		}
		// 如果没有设置 content-type 默认为 json
		if d.header == nil || d.header.Get(HeaderContentType) == "" {
			d.Type(MIMEApplicationJSON)
		}
	}
	req, err = http.NewRequest(d.method, url, r)
	if err != nil {
		return
	}
	// 如果有设置超时，则调整context
	if d.timeout != 0 {
		currentCtx := d.ctx
		if currentCtx == nil {
			currentCtx = context.Background()
		}
		ctx, cancel := context.WithTimeout(currentCtx, d.timeout)
		d.ctx = ctx
		d.OnDone(func(_ *Dusk) error {
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
	d.addAcceptEncoding(SnappyEncoding)
	d.OnResponse(SnappyDecode)
	return d
}

func (d *Dusk) do() (err error) {
	req := d.Request
	c := d.client
	if c == nil {
		c = http.DefaultClient
	}
	// 如果启用trace ，则添加相应的 context
	if d.enabledTrace {
		trace, ht := NewClientTrace()
		defer func() {
			ht.Done = time.Now()
		}()
		ctx := d.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		d.ctx = httptrace.WithClientTrace(ctx, trace)
		req = req.WithContext(d.ctx)
		d.Request = req
		d.ht = ht
	}
	err = d.EmitRequest(EventTypeBefore)
	if err != nil {
		return
	}
	resp, err := c.Do(req)
	if err != nil {
		return
	}
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
	defer resp.Body.Close()

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

	req, err := d.newReqest()
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
