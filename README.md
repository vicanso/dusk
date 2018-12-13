# dusk

[![Build Status](https://img.shields.io/travis/vicanso/dusk.svg?label=linux+build)](https://travis-ci.org/vicanso/dusk)

http request client support interceptor.

## API

```go
d := dusk.New()
d.Client = &http.Client{
  Timeout: 10 * time.Second,
}
d.ConvertError = func(err error, d *dusk.Dusk) error {
  // 可以根据需要转换为自定义的error对象
  newError := errors.New(err.Error())
  return newError
}
d.EnableTimelineTrace = true
d.On(dusk.EventResponse, func(d *dusk.Dusk) {
  // 可以根据需要将调整d.Response d.Error
  fmt.Println(d.Body)
})
d.On(dusk.EventDone, func(d *dusk.Dusk) {
  stats := d.GetTimelineStats()
  fmt.Println(stats)
})
resp, body, err := d.Get("https://aslant.site/", nil)
```

### Get

HTTP get request

```go
d := dusk.New()
resp, body, err := d.Get("https://aslant.site/", nil)
```

```go
d := dusk.New()
resp, body, err := d.Get("https://aslant.site/", map[string]string{
  "t": time.Now().Format(time.RFC3339),
})
```

### GetWithHeader

HTTP get request with http header

```go
d := dusk.New()
resp, body, err := d.GetWithHeader("https://aslant.site/", nil, nil)
```

```go
d := dusk.New()
resp, body, err := d.GetWithHeader("https://aslant.site/", map[string]string{
  "t": time.Now().Format(time.RFC3339),
}, map[string]string{
  "X-Token": "abc",
})
```

### Post

HTTP post request

```go
d := dusk.New()
resp, body, err := d.Post("https://aslant.site/users/login", map[string]string{
  "account": "tree.xie",
}, nil)
```

### PostWithHeader

HTTP post request with http header

```go
d := dusk.New()
resp, body, err := d.Post("https://aslant.site/users/login", map[string]string{
  "account": "tree.xie",
}, map[string]string{
  "t": time.Now().Format(time.RFC3339),
}, map[string]string{
  "X-Token": "abc",
})
```

### EnableTimelineTrace

Enable http timeline trace

```go
d := dusk.New()
d.EnableTimelineTrace = true
d.On(dusk.EventDone, func(d *dusk.Dusk) {
  buf, _ := json.Marshal(d.GetTimelineStats())
  // {"dnsLookup":2380544,"tcpConnection":58811605,"tlsHandshake":353453805,"serverProcessing":56606236,"contentTransfer":1776865,"total":474806674}
  fmt.Println(string(buf))
})
resp, body, err := d.Get("https://aslant.site/", nil)
```

### ConvertError

Convert the error to custom's error struct

```go
type CustomError struct {
	Message string
	Code    int
}
func (ce *CustomError) Error() string {
	return ce.Message
}

d := dusk.New()
d.ConvertError = func(err error, d *dusk.Dusk) error {
  ce := &CustomError{
    Message: err.Error(),
    Code:    -1,
  }
  return ce
}
resp, body, err := d.Get("https://aslant.site/", nil)
```

### Client

Change the default client for dusk

```go
d := dusk.New()
d.Client = &http.Client{
  Timeout: 3 * time.Second,
}
resp, body, err := d.Get("https://aslant.site/", nil)
```