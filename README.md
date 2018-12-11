# dusk

[![Build Status](https://img.shields.io/travis/vicanso/dusk.svg?label=linux+build)](https://travis-ci.org/vicanso/dusk)

http request client support interceptor.

## API

```go
d := dusk.New()
d.On(dusk.EventResponse, func(d *dusk.Dusk) {
  fmt.Println(d)
})
resp, body, err := d.Get("http://aslant.site/", nil)
```
