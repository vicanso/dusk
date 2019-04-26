# dusk

[![Build Status](https://img.shields.io/travis/vicanso/dusk.svg?label=linux+build)](https://travis-ci.org/vicanso/dusk)

Http request client supports interceptors, such as `OnRequest`, `OnRequestSuccess`, `OnResponse` and so on. It supports `br` and `snappy` content encoding.

## API

```go
d := dusk.Get("https://aslant.site/").Br()
// http client 尽量使用公共的实例，可以提高连接复用
d.SetClient(&http.Client{
  Timeout: 3 * time.Second,
})
d.EnableTrace()
// 请求发出前触发此事件
d.OnRequest(func(req *http.Request, d *dusk.Dusk) (newReq *http.Request, newErr error) {
  // 如果需要可以生成新的请求，则赋值至 newReq
  // 如果需要生成新的错误，则赋值至 newError，则请求出错返回
  return
})
// 当请求有响应时触发此事件
d.OnResponse(func(resp *http.Response, d *dusk.Dusk) (newResp *http.Response, newError error) {
  // 如果需要返回新的响应，则赋值至 newResp
  // 如果需要生成新的错误，则赋值至 newError，则请求出错返回
  return
})
// 无论请求成功或失败都会触发此事件
d.OnDone(func(d *dusk.Dusk) error {
  // 可增加一些系统统计等处理
  return nil
})

resp, body, err := d.Do()
fmt.Println(err)
fmt.Println(string(body))
fmt.Println(resp)
fmt.Println(d.GetHTTPTrace())
```

### Get

HTTP get request

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

### Post

HTTP Post request

```go
resp, body, err := dusk.Post("https://www.baidu.com/").
  Send(map[string]string{
    "foo": "bar",
  }).
  Timeout(3 * time.Second).
  Do()
fmt.Println(err)
fmt.Println(string(body))
fmt.Println(resp)
```