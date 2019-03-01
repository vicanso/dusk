package dusk

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/h2non/gock"
)

func TestGetURL(t *testing.T) {
	urlStr := "http://aslant.site/users/me?t=1&v=2"
	str := GetURL("http://aslant.site/users/me", map[string]string{
		"t": "1",
		"v": "2",
	})
	if str != urlStr {
		t.Fatalf("get url fail")
	}
	str = GetURL("http://aslant.site/users/me?t=1", map[string]string{
		"v": "2",
	})
	if str != urlStr {
		t.Fatalf("get url fail")
	}
}

func TestGet(t *testing.T) {
	defer gock.Off()
	t.Run("normal get", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, body, err := d.Get("http://aslant.site/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("get request fail, %d", resp.StatusCode)
		}

		if len(body) == 0 {
			t.Fatalf("get request body is empty")
		}
	})

	t.Run("url prefix", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := New()
		d.URLPrefix = "http://aslant.site"
		resp, _, err := d.Get("/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("get request fail, %d", resp.StatusCode)
		}
	})

	t.Run("get with header", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			MatchHeader("Token", "abc").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d := New()
		resp, _, err := d.GetWithHeader("http://aslant.site/", nil, map[string]string{
			"Token": "abc",
		})
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("get request fail, %d", resp.StatusCode)
		}
	})
}

func TestPost(t *testing.T) {
	defer gock.Off()
	t.Run("normal post", func(t *testing.T) {
		gock.New("http://aslant.site").
			Post("/").
			MatchType("json").
			JSON(map[string]string{"foo": "bar"}).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.Post("http://aslant.site/", map[string]string{
			"foo": "bar",
		}, nil)
		if err != nil {
			t.Fatalf("post request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("post request fail, %d", resp.StatusCode)
		}
	})

	t.Run("post with header", func(t *testing.T) {
		gock.New("http://aslant.site").
			Post("/").
			MatchHeader("Token", "abc").
			MatchType("json").
			JSON(map[string]string{"foo": "bar"}).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.PostWithHeader("http://aslant.site/", map[string]string{
			"foo": "bar",
		}, nil, map[string]string{
			"Token": "abc",
		})
		if err != nil {
			t.Fatalf("post request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("post request fail, %d", resp.StatusCode)
		}
	})
}

func TestNewRequest(t *testing.T) {
	d := Dusk{}
	data := bytes.NewReader([]byte(`{
		"foo": "bar"
	}`))
	d.Timeout = time.Second
	req, err := d.NewRequest("POST", "http://aslant.site/", nil, data, nil)
	_, ok := req.Context().Deadline()
	if !ok {
		t.Fatalf("set request timeout fail")
	}
	if err != nil {
		t.Fatalf("new request fail, %v", err)
	}
	if req.Header.Get(HeaderContentType) != "" {
		t.Fatalf("request content type should be nil")
	}
}

func TestGzipResponse(t *testing.T) {
	resBody := []byte(`{"name":"tree.xie"}`)
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(resBody)
	w.Close()

	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(b.Bytes()).
		Header.Set(contentEncoding, gzipEncoding)

	d := New()
	resp, body, err := d.Get("http://aslant.site/", nil)
	if err != nil {
		t.Fatalf("get request fail, %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("get request fail, %d", resp.StatusCode)
	}
	if !bytes.Equal(resBody, body) {
		t.Fatalf("response body is invalid")
	}
}

func TestDel(t *testing.T) {
	defer gock.Off()
	t.Run("normal del", func(t *testing.T) {
		gock.New("http://aslant.site").
			Delete("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.Del("http://aslant.site/", map[string]string{
			"v": "1",
		})
		if err != nil {
			t.Fatalf("del request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("del request fail, %d", resp.StatusCode)
		}
	})
	t.Run("del with header", func(t *testing.T) {
		gock.New("http://aslant.site").
			Delete("/").
			MatchHeader("Token", "abc").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.DelWithHeader("http://aslant.site/", map[string]string{
			"v": "1",
		}, map[string]string{
			"Token": "abc",
		})
		if err != nil {
			t.Fatalf("del request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("del request fail, %d", resp.StatusCode)
		}
	})
}

func TestPatch(t *testing.T) {
	defer gock.Off()
	t.Run("normal patch", func(t *testing.T) {
		gock.New("http://aslant.site").
			Patch("/").
			MatchType("json").
			JSON(map[string]string{"foo": "bar"}).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.Patch("http://aslant.site/", map[string]string{
			"foo": "bar",
		}, nil)
		if err != nil {
			t.Fatalf("patch request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("patch request fail, %d", resp.StatusCode)
		}
	})
	t.Run("patch with header", func(t *testing.T) {
		gock.New("http://aslant.site").
			Patch("/").
			MatchHeader("Token", "abc").
			MatchType("json").
			JSON(map[string]string{"foo": "bar"}).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.PatchWithHeader("http://aslant.site/", map[string]string{
			"foo": "bar",
		}, nil, map[string]string{
			"Token": "abc",
		})
		if err != nil {
			t.Fatalf("patch request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("patch request fail, %d", resp.StatusCode)
		}
	})
}

func TestPut(t *testing.T) {
	defer gock.Off()
	t.Run("normal put", func(t *testing.T) {
		gock.New("http://aslant.site").
			Put("/").
			MatchType("json").
			JSON(map[string]string{"foo": "bar"}).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.Put("http://aslant.site/", map[string]string{
			"foo": "bar",
		}, nil)
		if err != nil {
			t.Fatalf("put request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("put request fail, %d", resp.StatusCode)
		}
	})
	t.Run("put with header", func(t *testing.T) {
		gock.New("http://aslant.site").
			Put("/").
			MatchHeader("Token", "abc").
			MatchType("json").
			JSON(map[string]string{"foo": "bar"}).
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		resp, _, err := d.PutWithHeader("http://aslant.site/", map[string]string{
			"foo": "bar",
		}, nil, map[string]string{
			"Token": "abc",
		})
		if err != nil {
			t.Fatalf("put request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("put request fail, %d", resp.StatusCode)
		}
	})
}

func TestEnableTimelineTrace(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	d := New()
	d.EnableTimelineTrace = true
	d.On(EventDone, func(d *Dusk) {
		stats := d.GetTimelineStats()
		if stats == nil || stats.Total.Nanoseconds() == 0 {
			t.Fatalf("get timeline stats fail")
		}
	})
	resp, body, err := d.Get("http://aslant.site/", nil)
	if err != nil {
		t.Fatalf("get request fail, %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("get request fail, %d", resp.StatusCode)
	}

	if len(body) == 0 {
		t.Fatalf("get request body is empty")
	}
}

func TestOnEvent(t *testing.T) {
	defer gock.Off()
	t.Run("event step", func(t *testing.T) {

		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		step := 0
		d.On(EventRequest, func(_ *Dusk) {
			step++
			if step != 1 {
				t.Fatalf("the request event should be step 1")
			}
		})
		d.On(EventResponse, func(_ *Dusk) {
			step++
			if step != 2 {
				t.Fatalf("the response event should be step 2")
			}
		})
		d.On(EventDone, func(_ *Dusk) {
			step++
			if step != 3 {
				t.Fatalf("the done event should be step 3")
			}
		})

		resp, body, err := d.Get("http://aslant.site/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("get request fail, %d", resp.StatusCode)
		}

		if len(body) == 0 {
			t.Fatalf("get request body is empty")
		}
	})

	t.Run("modify url on request event", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		d.On(EventRequest, func(d *Dusk) {
			// 填充请求的 host schema
			d.Request.URL.Host = "aslant.site"
			d.Request.URL.Scheme = "http"
		})
		resp, body, err := d.Get("/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("get request fail, %d", resp.StatusCode)
		}

		if len(body) == 0 {
			t.Fatalf("get request body is empty")
		}
	})

	t.Run("error event", func(t *testing.T) {
		d := New()
		done := false
		d.On(EventError, func(_ *Dusk) {
			if d.Error != nil {
				done = true
			}
		})
		_, _, err := d.Get("http://aslant.site/", nil)
		if err == nil || !done {
			t.Fatalf("error event is not emitted")
		}
	})

	t.Run("create an error on request event", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		customError := errors.New("abc")
		d := New()
		done := false
		d.On(EventRequest, func(d *Dusk) {
			d.Error = customError
		})
		d.On(EventError, func(_ *Dusk) {
			done = true
		})
		_, _, err := d.Get("http://aslant.site/", nil)
		if err != customError {
			t.Fatalf("create custom error on event fail")
		}
		if !done {
			t.Fatalf("miss error event")
		}
	})

	t.Run("create an error on response event", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		customError := errors.New("abc")
		d := New()
		done := false
		d.On(EventResponse, func(d *Dusk) {
			d.Error = customError
		})
		d.On(EventError, func(_ *Dusk) {
			done = true
		})
		_, _, err := d.Get("http://aslant.site/", nil)
		if err != customError {
			t.Fatalf("create custom error on event fail")
		}
		if !done {
			t.Fatalf("miss error event")
		}
	})
}

func TestConvertError(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(500).
		JSON(map[string]string{
			"message": "error",
		})
	d := New()

	d.On(EventResponse, func(d *Dusk) {
		status := d.Response.Status
		if status != "200" {
			d.Error = errors.New("status is:" + status)
		}
	})
	d.ConvertError = func(err error, _ *Dusk) error {
		return errors.New(err.Error() + " covert")
	}
	_, _, err := d.Get("http://aslant.site/", nil)
	if err.Error() != "status is:500 Internal Server Error covert" {
		t.Fatalf("convert error fail")
	}
}

func TestGetSetValue(t *testing.T) {
	d := New()
	key := "name"
	value := "tree.xie"
	if d.GetValue(key) != nil {
		t.Fatalf("get value should be nil before set")
	}
	d.SetValue(key, value)
	v := d.GetValue(key).(string)
	if v != value {
		t.Fatalf("get set value fail")
	}
}

func TestReset(t *testing.T) {
	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})
	d := New()
	d.Get("http://aslant.site/", nil)
	d.Reset()
	if d.Request != nil ||
		d.Response != nil ||
		d.Error != nil ||
		d.M != nil {
		t.Fatalf("reset fail")
	}
}

func TestContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	d := New()
	d.SetContext(ctx)
	if d.GetContext() != ctx {
		t.Fatalf("set context fail")
	}
}
