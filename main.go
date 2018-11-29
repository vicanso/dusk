package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vicanso/dusk/dusk"
)

func main() {
	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   3 * time.Second,
	}
	d := dusk.New()
	// 如果不设置使用默认client
	d.Client = client
	d.EnableTimelineTrace = true
	d.On(dusk.EventDone, func(d *dusk.Dusk) {
		fmt.Println(d.GetTimelineStats())
	})

	resp, body, err := d.Get("https://www.baidu.com/", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	fmt.Println(resp)
}
