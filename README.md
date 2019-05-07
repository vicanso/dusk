# dusk

[![Build Status](https://img.shields.io/travis/vicanso/dusk.svg?label=linux+build)](https://travis-ci.org/vicanso/dusk)

Http request client supports interceptors, such as `OnRequest`, `OnRequestSuccess`, `OnResponse` and so on. It supports `br` and `snappy` content encoding.

## API

```go
dusk.AddRequestListener(func(req *http.Request, d *dusk.Dusk) (newErr error) {
  fmt.Println("global request event")
  return
}, dusk.EventTypeBefore)
dusk.AddResponseListener(func(resp *http.Response, d *dusk.Dusk) (newError error) {
  fmt.Println("global response event")
  return
}, dusk.EventTypeBefore)

d := dusk.Get("https://aslant.site/").Br()
// http client 尽量使用公共的实例，可以提高连接复用
d.SetClient(&http.Client{
  Timeout: 3 * time.Second,
})
d.EnableTrace()
// 请求发出前触发此事件
d.AddRequestListener(func(req *http.Request, d *dusk.Dusk) (newErr error) {
  fmt.Println("before request event")
  // 如果需要可以生成新的请求，则赋值至 newReq
  // 如果需要生成新的错误，则赋值至 newError，则请求出错返回
  return
}, dusk.EventTypeBefore)
// 当请求有响应时触发此事件
d.AddResponseListener(func(resp *http.Response, d *dusk.Dusk) (newError error) {
  fmt.Println("before response event")
  // 如果需要返回新的响应，则赋值至 newResp
  // 如果需要生成新的错误，则赋值至 newError，则请求出错返回
  return
}, dusk.EventTypeBefore)
// 无论请求成功或失败都会触发此事件
d.AddDoneListener(func(d *dusk.Dusk) error {
  fmt.Println("done event")
  // 可增加一些系统统计等处理
  return nil
})

resp, body, err := d.Do()
fmt.Println(err)
fmt.Println(len(body))
fmt.Println(resp)
fmt.Println(d.GetHTTPTrace())
```

### Get/Post/Put/Patch/Delete

Do http request

```go
resp, body, err := dusk.Get("https://www.baidu.com/").
  // set http request header
  Set("X-Request-ID", "1234").
  // set http request query
  Query("type", "vip").
  Queries(map[string]string{
    "site": "mobile",
  }).
  Timeout(3 * time.Second).
  Do()
fmt.Println(err)
fmt.Println(string(body))
fmt.Println(resp)
```

### NewInstance

Create an http request instance, it support http requsets.

```go
ins := dusk.NewInstance()

ins.AddRequestListener(func(req *http.Request, _ *dusk.Dusk) (newErr error) {
  req.URL.Scheme = "https"
  req.URL.Host = "www.baidu.com"
  return
}, dusk.EventTypeBefore)

resp, _, err := ins.Get("/").Do()
fmt.Println(resp)
fmt.Println(err)
```
