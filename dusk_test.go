package dusk

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/snappy"
	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestSetClient(t *testing.T) {

	d := Dusk{}
	client := &http.Client{}
	d.SetClient(client)
	assert.Equal(t, d.GetClient(), client)
}

func TestSetGetValue(t *testing.T) {
	d := &Dusk{}
	d.SetValue("a", 1)
	assert.Equal(t, d.GetValue("a").(int), 1)
}

func TestSetGetContext(t *testing.T) {
	d := &Dusk{}
	ctx := context.Background()
	d.SetContext(ctx)
	assert.Equal(t, d.ctx, ctx)
	assert.Equal(t, d.GetContext(), ctx)
}

func TestHTTPGet(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	d := Get("http://aslant.site/")
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
}

func TestHTTPHead(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Head("/").
		Reply(200)

	d := Head("http://aslant.site/")
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(len(body), 0)
}

func TestHTTPPut(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Put("/").
		Reply(200)

	d := Put("http://aslant.site/")
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(len(body), 0)
}

func TestHTTPPatch(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Patch("/").
		Reply(200)

	d := Patch("http://aslant.site/")
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(len(body), 0)
}

func TestHTTPDelete(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Delete("/").
		Reply(200)

	d := Delete("http://aslant.site/")
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(len(body), 0)
}

func TestHTTPPost(t *testing.T) {
	t.Run("post json", func(t *testing.T) {
		assert := assert.New(t)
		defer gock.Off()
		gock.New("http://aslant.site").
			Post("/123").
			BodyString(`{"account":"tree.xie"}`).
			MatchHeader("a", "1").
			MatchHeader("Content-Type", "application/json").
			MatchParam("type", "2").
			MatchParam("category", "3").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := Post("http://aslant.site/:id").
			Param("id", "123").
			Send(map[string]string{
				"account": "tree.xie",
			}).
			Set("a", "1").
			Queries(map[string]string{
				"type": "2",
			}).
			Query("category", "3")

		resp, body, err := d.Do()
		assert.Nil(err)
		assert.Equal(resp.StatusCode, 200)
		assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
	})

	t.Run("post form", func(t *testing.T) {
		data := make(url.Values)
		data.Add("type", "1")
		data.Add("type", "2")
		data.Set("account", "tree.xie")
		assert := assert.New(t)
		defer gock.Off()
		gock.New("http://aslant.site").
			Post("/123").
			MatchHeader("Content-Type", "application/x-www-form-urlencoded").
			BodyString(`account=tree.xie&type=1&type=2`).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := Post("http://aslant.site/:id").
			Param("id", "123").
			Send(data)
		resp, body, err := d.Do()
		assert.Nil(err)
		assert.Equal(resp.StatusCode, 200)
		assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
	})
}

func TestTimeout(t *testing.T) {
	assert := assert.New(t)
	d := Get("https://aslant.site/").
		EnableTrace().
		Timeout(time.Millisecond)
	_, _, err := d.Do()
	ue, ok := err.(*url.Error)
	assert.True(ok)
	assert.True(ue.Timeout())
}

func TestEvent(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	requestURI := "http://aslant.site/?a=1&b=2"
	requestEventBefore := false
	requestEventSuccess := false
	responseEventBefore := false
	responseEventSuccess := false
	doneEvent := false

	d := Get(requestURI)
	d.AddRequestListener(func(req *http.Request, _ *Dusk) (newReq *http.Request, err error) {
		assert.Equal(req.URL.String(), requestURI)
		requestEventBefore = true
		return
	}, EventTypeBefore)

	d.AddRequestListener(func(req *http.Request, _ *Dusk) (newReq *http.Request, err error) {
		if requestEventBefore {
			requestEventSuccess = true
		}
		return
	}, EventTypeAfter)
	d.AddResponseListener(func(resp *http.Response, _ *Dusk) (newResp *http.Response, err error) {
		responseEventBefore = true
		return
	}, EventTypeBefore)
	d.AddResponseListener(func(resp *http.Response, _ *Dusk) (newResp *http.Response, err error) {
		if responseEventBefore {
			responseEventSuccess = true
		}
		return
	}, EventTypeAfter)

	d.AddDoneListener(func(_ *Dusk) (err error) {
		doneEvent = true
		return
	})

	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
	assert.True(requestEventBefore)
	assert.True(requestEventSuccess)
	assert.True(responseEventBefore)
	assert.True(responseEventSuccess)
	assert.True(doneEvent)
}

func TestResponseBodySnappy(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	var dst []byte
	buf := snappy.Encode(dst, []byte(`{"name":"tree.xie"}`))

	gock.New("http://aslant.site").
		Get("/").
		MatchHeader(HeaderAcceptEncoding, GzipEncoding+", "+SnappyEncoding).
		Reply(200).
		SetHeader(HeaderContentEncoding, SnappyEncoding).
		SetHeader(HeaderContentLength, strconv.Itoa(len(buf))).
		Body(bytes.NewReader(buf))

	d := Get("http://aslant.site/").
		Snappy()
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
	assert.Equal(resp.Header.Get(HeaderContentLength), "")
}

func TestResponseBodyBrotli(t *testing.T) {
	assert := assert.New(t)
	// abcd的br压缩
	brBase64 := "iwGAYWJjZAM="

	defer gock.Off()
	buf, err := base64.StdEncoding.DecodeString(brBase64)
	assert.Nil(err)
	gock.New("http://aslant.site").
		Get("/").
		MatchHeader(HeaderAcceptEncoding, GzipEncoding+", "+BrEncoding).
		Reply(200).
		SetHeader(HeaderContentEncoding, BrEncoding).
		SetHeader(HeaderContentLength, strconv.Itoa(len(buf))).
		Body(bytes.NewReader(buf))

	d := Get("http://aslant.site/").
		Br()
	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(strings.TrimSpace(string(body)), "abcd")
	assert.Equal(resp.Header.Get(HeaderContentLength), "")
}

func TestEnableTrace(t *testing.T) {
	assert := assert.New(t)
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
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
	assert.NotNil(d.GetHTTPTrace())
}

func TestEmitRequest(t *testing.T) {
	defer gock.Off()

	t.Run("new request", func(t *testing.T) {
		assert := assert.New(t)
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		r := httptest.NewRequest("GET", "/users/me", nil)
		d := Get("http://aslant.site/")
		d.AddRequestListener(func(_ *http.Request, _ *Dusk) (newReq *http.Request, err error) {
			newReq = r
			return
		}, EventTypeBefore)
		// 不判断是否出错，只需要后面检查request 是否被替换
		d.Do()
		assert.Equal(d.Request, r)
	})

	t.Run("return error", func(t *testing.T) {
		assert := assert.New(t)
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		e := errors.New("abcd")
		d := Get("http://aslant.site/")
		d.AddRequestListener(func(_ *http.Request, _ *Dusk) (newReq *http.Request, err error) {
			err = e
			return
		}, EventTypeBefore)
		_, _, err := d.Do()
		assert.Equal(err, e)
	})
}

func TestEmitResponse(t *testing.T) {
	defer gock.Off()
	t.Run("new response", func(t *testing.T) {
		assert := assert.New(t)
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := Get("http://aslant.site/")
		d.AddResponseListener(func(_ *http.Response, _ *Dusk) (newResp *http.Response, err error) {
			newResp = &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(`{"name":"abcd"}`))),
			}
			return
		}, EventTypeAfter)
		resp, body, err := d.Do()
		assert.Nil(err)
		assert.Equal(resp.StatusCode, 200)
		assert.Equal(strings.TrimSpace(string(body)), `{"name":"abcd"}`)
	})

	t.Run("read body by custom", func(t *testing.T) {
		assert := assert.New(t)
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := Get("http://aslant.site/")
		d.AddResponseListener(func(_ *http.Response, d *Dusk) (newResp *http.Response, err error) {
			d.Body = []byte(`{"name":"abcd"}`)
			return
		}, EventTypeBefore)
		resp, body, err := d.Do()
		assert.Nil(err)
		assert.Equal(resp.StatusCode, 200)
		assert.Equal(strings.TrimSpace(string(body)), `{"name":"abcd"}`)
	})

	t.Run("return error", func(t *testing.T) {
		assert := assert.New(t)
		e := errors.New("abcd")
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := Get("http://aslant.site/")
		d.AddResponseListener(func(_ *http.Response, d *Dusk) (newResp *http.Response, err error) {
			err = e
			return
		}, EventTypeBefore)
		_, _, err := d.Do()
		assert.Equal(err, e)
	})
}

func TestSetType(t *testing.T) {
	assert := assert.New(t)
	d := Post("/users/me")
	d.Type("json")
	assert.Equal(d.header.Get(HeaderContentType), MIMEApplicationJSON)
	d.Type("form")
	assert.Equal(d.header.Get(HeaderContentType), MIMEApplicationFormUrlencoded)
}

func TestEmitError(t *testing.T) {
	defer ClearErrorListener()
	globalErrorDone := false
	AddErrorListener(func(_ error, _ *Dusk) error {
		globalErrorDone = true
		return nil
	})
	assert := assert.New(t)
	e := errors.New("abcd")
	d := Get("http://aslant.site/")
	d.AddErrorListener(func(err error, _ *Dusk) (newErr error) {
		assert.True(globalErrorDone)
		newErr = e
		return
	})
	d.Timeout(time.Nanosecond)
	_, _, err := d.Do()
	assert.Equal(err, e)
}

func TestIsDisableCompression(t *testing.T) {
	assert := assert.New(t)
	d := new(Dusk)
	assert.False(d.isDisableCompression())
	d.SetClient(&http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	})
	assert.True(d.isDisableCompression())
}

func TestGetAttr(t *testing.T) {
	assert := assert.New(t)
	d := Get("/:id")
	assert.Equal(d.GetMethod(), "GET")
	assert.Equal(d.GetPath(), "/:id")
}

func TestGlobalEvent(t *testing.T) {
	defer ClearRequestListener()
	defer ClearResponseListener()
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	requestURI := "http://aslant.site/?a=1&b=2"
	requestEventBefore := false
	requestEventSuccess := false
	responseEventBefore := false
	responseEventSuccess := false
	doneEvent := false
	AddRequestListener(func(req *http.Request, _ *Dusk) (newReq *http.Request, err error) {
		assert.Equal(req.URL.String(), requestURI)
		requestEventBefore = true
		return
	}, EventTypeBefore)

	AddRequestListener(func(req *http.Request, _ *Dusk) (newReq *http.Request, err error) {
		if requestEventBefore {
			requestEventSuccess = true
		}
		return
	}, EventTypeAfter)
	AddResponseListener(func(resp *http.Response, _ *Dusk) (newResp *http.Response, err error) {
		responseEventBefore = true
		return
	}, EventTypeBefore)
	AddResponseListener(func(resp *http.Response, _ *Dusk) (newResp *http.Response, err error) {
		if responseEventBefore {
			responseEventSuccess = true
		}
		return
	}, EventTypeAfter)

	d := Get(requestURI)

	d.AddDoneListener(func(_ *Dusk) (err error) {
		doneEvent = true
		return
	})

	resp, body, err := d.Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.Equal(strings.TrimSpace(string(body)), `{"name":"tree.xie"}`)
	assert.True(requestEventBefore)
	assert.True(requestEventSuccess)
	assert.True(responseEventBefore)
	assert.True(responseEventSuccess)
	assert.True(doneEvent)
}
