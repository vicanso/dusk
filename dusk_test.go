package dusk

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
)

func TestSetClient(t *testing.T) {
	d := Dusk{}
	client := &http.Client{}
	d.SetClient(client)
	if d.client != client {
		t.Fatalf("set client fail")
	}
}

func TestSetGetValue(t *testing.T) {
	d := &Dusk{}
	d.SetValue("a", 1)
	if d.GetValue("a").(int) != 1 {
		t.Fatalf("set/get value fail")
	}
}

func TestSetGetContext(t *testing.T) {
	d := &Dusk{}
	ctx := context.Background()
	d.SetContext(ctx)
	if d.ctx != ctx || d.GetContext() != ctx {
		t.Fatalf("set/get context fail")
	}
}

func TestHTTPGet(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	d := Get("http://aslant.site/")
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("get request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		strings.TrimSpace(string(body)) != `{"name":"tree.xie"}` {
		t.Fatalf("response of get request invalid")
	}
}

func TestHTTPHead(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Head("/").
		Reply(200)

	d := Head("http://aslant.site/")
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("head request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		len(body) != 0 {
		t.Fatalf("response of head request invalid")
	}
}

func TestHTTPPut(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Put("/").
		Reply(200)

	d := Put("http://aslant.site/")
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("put request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		len(body) != 0 {
		t.Fatalf("response of put request invalid")
	}
}

func TestHTTPPatch(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Patch("/").
		Reply(200)

	d := Patch("http://aslant.site/")
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("patch request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		len(body) != 0 {
		t.Fatalf("response of patch request invalid")
	}
}

func TestHTTPDelete(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Delete("/").
		Reply(200)

	d := Delete("http://aslant.site/")
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("delete request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		len(body) != 0 {
		t.Fatalf("response of delete request invalid")
	}
}

func TestHTTPPost(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Post("/").
		BodyString(`{"account":"tree.xie"}`).
		MatchHeader("a", "1").
		MatchParam("type", "2").
		MatchParam("category", "3").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	d := Post("http://aslant.site/").
		Send(map[string]string{
			"account": "tree.xie",
		}).
		Set("a", "1").
		Queries(map[string]string{
			"type": "2",
		}).
		Query("category", "3")

	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("post request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		strings.TrimSpace(string(body)) != `{"name":"tree.xie"}` {
		t.Fatalf("response of post request invalid")
	}
}

func TestTimeout(t *testing.T) {
	d := Get("https://aslant.site/").
		Timeout(time.Nanosecond)
	_, _, err := d.Do()
	ue, ok := err.(*url.Error)
	if !ok || !ue.Timeout() {
		t.Fatalf("set request timeout fail")
	}
}

func TestEvent(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	requestURI := "http://aslant.site/?a=1&b=2"
	requestEvent := false
	responseEvent := false
	doneEvent := false

	d := Get(requestURI)
	d.OnRequest(func(req *http.Request, _ *Dusk) (newReq *http.Request, err error) {
		if req.URL.String() != requestURI {
			t.Fatalf("request uri invalid")
		}
		requestEvent = true
		return
	})
	d.OnResponse(func(resp *http.Response, _ *Dusk) (newResp *http.Response, err error) {
		responseEvent = true
		return
	})
	d.OnDone(func(_ *Dusk) (err error) {
		doneEvent = true
		return
	})

	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("get request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		strings.TrimSpace(string(body)) != `{"name":"tree.xie"}` {
		t.Fatalf("response of get request invalid")
	}
	if !requestEvent ||
		!responseEvent ||
		!doneEvent {
		t.Fatalf("not all event was emitted")
	}
}

func TestResponseBodyGzip(t *testing.T) {
	defer gock.Off()
	var b bytes.Buffer
	w, _ := gzip.NewWriterLevel(&b, 1)
	w.Write([]byte(`{"name":"tree.xie"}`))
	w.Close()

	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		SetHeader(HeaderContentEncoding, GzipEncoding).
		Body(bytes.NewReader(b.Bytes()))

	d := Get("http://aslant.site/")
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("get request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		strings.TrimSpace(string(body)) != `{"name":"tree.xie"}` {
		t.Fatalf("gzip response of get request invalid")
	}
}

func TestEnableTrace(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	d := Get("http://aslant.site/")
	d.EnableTrace()
	resp, body, err := d.Do()
	if err != nil {
		t.Fatalf("get request fail, %v", err)
	}
	if resp.StatusCode != 200 ||
		strings.TrimSpace(string(body)) != `{"name":"tree.xie"}` {
		t.Fatalf("response of get request invalid")
	}
	if d.GetHTTPTrace() == nil {
		t.Fatalf("enable trace fail")
	}
}

func TestEmitRequest(t *testing.T) {
	defer gock.Off()

	t.Run("new request", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		r := httptest.NewRequest("GET", "/users/me", nil)
		d := Get("http://aslant.site/")
		d.OnRequest(func(_ *http.Request, _ *Dusk) (newReq *http.Request, err error) {
			newReq = r
			return
		})
		// 不判断是否出错，只需要后面检查request 是否被替换
		d.Do()
		if d.Request != r {
			t.Fatalf("convert new request fail")
		}
	})

	t.Run("return error", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		e := errors.New("abcd")
		d := Get("http://aslant.site/")
		d.OnRequest(func(_ *http.Request, _ *Dusk) (newReq *http.Request, err error) {
			err = e
			return
		})
		_, _, err := d.Do()
		if err != e {
			t.Fatalf("on request event return error fail")
		}
	})
}

func TestEmitResponse(t *testing.T) {
	defer gock.Off()
	t.Run("new response", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := Get("http://aslant.site/")
		d.OnResponse(func(_ *http.Response, _ *Dusk) (newResp *http.Response, err error) {
			newResp = &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(`{"name":"abcd"}`))),
			}
			return
		})
		resp, body, err := d.Do()
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 ||
			strings.TrimSpace(string(body)) != `{"name":"abcd"}` {
			t.Fatalf("response of get request invalid")
		}
	})

	t.Run("read body by custom", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := Get("http://aslant.site/")
		d.OnResponse(func(_ *http.Response, d *Dusk) (newResp *http.Response, err error) {
			d.Body = []byte(`{"name":"abcd"}`)
			return
		})
		resp, body, err := d.Do()
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 ||
			strings.TrimSpace(string(body)) != `{"name":"abcd"}` {
			t.Fatalf("response of get request invalid")
		}
	})

	t.Run("return error", func(t *testing.T) {
		e := errors.New("abcd")
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := Get("http://aslant.site/")
		d.OnResponse(func(_ *http.Response, d *Dusk) (newResp *http.Response, err error) {
			err = e
			return
		})
		_, _, err := d.Do()
		if err != e {
			t.Fatalf("on response event return error fail")
		}
	})
}

func TestSetType(t *testing.T) {
	d := Post("/users/me")
	d.Type("json")
	if d.header.Get(HeaderContentType) != MIMEApplicationJSON {
		t.Fatalf("set content-type: json fail")
	}
	d.Type("form")
	if d.header.Get(HeaderContentType) != MIMEApplicationFormUrlencoded {
		t.Fatalf("set content-type: form fail")
	}
}

func TestEmitError(t *testing.T) {
	e := errors.New("abcd")
	d := Get("http://aslant.site/")
	d.OnError(func(err error, _ *Dusk) (newErr error) {
		newErr = e
		return
	})
	d.Timeout(time.Nanosecond)
	_, _, err := d.Do()
	if err != e {
		t.Fatalf("on error event return new error fail")
	}
}
