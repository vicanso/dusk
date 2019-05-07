package dusk

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestInstance(t *testing.T) {
	assert := assert.New(t)
	ins := NewInstance()
	url := "http://aslant.site/"

	d := ins.Get(url)
	assert.Equal(d.method, "GET")

	d = ins.Head(url)
	assert.Equal(d.method, "HEAD")

	d = ins.Post(url)
	assert.Equal(d.method, "POST")

	d = ins.Put(url)
	assert.Equal(d.method, "PUT")

	d = ins.Patch(url)
	assert.Equal(d.method, "PATCH")

	d = ins.Delete(url)
	assert.Equal(d.method, "DELETE")
}

func TestInstanceEvent(t *testing.T) {
	assert := assert.New(t)
	ins := NewInstance()
	requestBeforeDone := false
	requestAfterDone := false
	responseBeforeDone := false
	responseAfterDone := false

	ins.AddRequestListener(func(req *http.Request, _ *Dusk) (err error) {
		requestBeforeDone = true
		return
	}, EventTypeBefore)

	ins.AddRequestListener(func(req *http.Request, _ *Dusk) (err error) {
		requestAfterDone = true
		return
	}, EventTypeAfter)

	ins.AddResponseListener(func(resp *http.Response, _ *Dusk) (err error) {
		responseBeforeDone = true
		return
	}, EventTypeBefore)

	ins.AddResponseListener(func(resp *http.Response, _ *Dusk) (err error) {
		responseAfterDone = true
		return
	}, EventTypeAfter)

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

	d := ins.Post("http://aslant.site/:id").
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
	assert.True(requestBeforeDone)
	assert.True(requestAfterDone)
	assert.True(responseBeforeDone)
	assert.True(responseAfterDone)
}

func TestInstanceErrorListener(t *testing.T) {
	assert := assert.New(t)
	ins := NewInstance()
	errorListenerDone := false
	ins.AddErrorListener(func(err error, _ *Dusk) error {
		errorListenerDone = true
		return nil
	})
	_, _, err := ins.Get("http://abc/").Do()
	assert.NotNil(err)
	assert.True(errorListenerDone)
}

func TestInstanceDoneListener(t *testing.T) {

	defer gock.Off()
	gock.New("http://aslant.site").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"name": "tree.xie",
		})

	assert := assert.New(t)
	ins := NewInstance()

	doneListenerDone := false
	ins.AddDoneListener(func(_ *Dusk) error {
		doneListenerDone = true
		return nil
	})

	resp, _, err := ins.Get("http://aslant.site/").Do()
	assert.Nil(err)
	assert.Equal(resp.StatusCode, 200)
	assert.True(doneListenerDone)
}
