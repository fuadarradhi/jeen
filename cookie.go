package jeen

import "net/http"

// Cookie is response and request for http cookie
type Cookie struct {
	request *http.Request
	writer  http.ResponseWriter
}

// create new cookie instance
func newCookie(rw http.ResponseWriter, r *http.Request) *Cookie {
	return &Cookie{
		writer:  rw,
		request: r,
	}
}

// Get returns the named cookie provided in the request or
// ErrNoCookie if not found.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (c *Cookie) Get(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

// AddCookie adds a cookie to the request. Per RFC 6265 section 5.4,
// AddCookie does not attach more than one Cookie header field. That
// means all cookies, if any, are written into the same line,
// separated by semicolon.
// AddCookie only sanitizes c's name and value, and does not sanitize
// a Cookie header already present in the request.
func (c *Cookie) Add(cookie *http.Cookie) {
	c.request.AddCookie(cookie)
}

// Set adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func (c *Cookie) Set(cookie *http.Cookie) {
	http.SetCookie(c.writer, cookie)
}

// All parses and returns the HTTP cookies sent with the request.
func (c *Cookie) All() []*http.Cookie {
	return c.request.Cookies()
}
