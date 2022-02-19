package jeen

import "net/http"

type Cookie struct {
	request *http.Request
	writer  http.ResponseWriter
}

func (c *Cookie) Get(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

func (c *Cookie) Set(cookie *http.Cookie) {
	http.SetCookie(c.writer, cookie)
}

func (c *Cookie) All() []*http.Cookie {
	return c.request.Cookies()
}
