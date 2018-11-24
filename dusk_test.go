package dusk

import (
	"errors"
	"net/http"
	"net/url"
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

func TestAddBeforeRequset(t *testing.T) {
	defer gock.Off()
	t.Run("add global before request hook", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		localhost := "localhost"
		mockErr := errors.New("abcd")
		AddBeforeRequest(func(_ *Dusk, req *http.Request) (err error) {
			if req.URL.RequestURI() == "/users/me" {
				return mockErr
			}
			// change url for localhost
			if req.Host != localhost {
				return
			}
			req.URL, _ = url.Parse("http://aslant.site/")
			req.Host = "aslant.site"
			return
		})

		d := New()
		resp, body, err := d.Get("http://"+localhost+"/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if resp.StatusCode != 200 || len(body) == 0 {
			t.Fatalf("change request url fail")
		}

		d = New()
		_, _, err = d.Get("http://localhost/users/me", nil)
		if err != mockErr {
			t.Fatalf("hook return error fail")
		}
	})

	t.Run("add before request hook", func(t *testing.T) {
		defer gock.Off()

		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		startedAt := "startedAt"
		now := time.Now().UnixNano()
		d.AddBeforeRequest(func(d *Dusk, req *http.Request) (err error) {
			d.SetValue(startedAt, now)
			return
		})
		_, _, err := d.Get("http://aslant.site/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		v := d.GetValue(startedAt).(int64)
		if v != now {
			t.Fatalf("get/set value fail")
		}
	})
}

func TestAddBeforeResponse(t *testing.T) {
	defer gock.Off()
	t.Run("add global before response hook", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		done := false
		AddBeforeResponse(func(d *Dusk, resp *http.Response) (err error) {
			done = true
			return
		})
		d := New()
		_, _, err := d.Get("http://aslant.site/", nil)
		if err != nil {
			t.Fatalf("get request fail, %v", err)
		}
		if !done {
			t.Fatalf("add before response hook fail")
		}
	})

	t.Run("add before response hook", func(t *testing.T) {
		gock.New("http://aslant.site").
			Get("/").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})

		d := New()
		mockErr := errors.New("abcd")
		d.AddBeforeResponse(func(d *Dusk, resp *http.Response) (err error) {
			return mockErr
		})
		_, _, err := d.Get("http://aslant.site/", nil)
		if err != mockErr {
			t.Fatalf("hook return error fail")
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
	d.On(EventResponse, func(d *Dusk) {
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
