package jeen

import (
	"context"
	"time"

	"github.com/alexedwards/scs/v2"
)

// SCS Session wrapper
// credit: github.com/alexedwards/scs/v2
type Session struct {
	session *scs.SessionManager
	context context.Context
}

// session value returned from Get
type sessionValue struct {
	value interface{}
}

// Create wrapper for scs session manager.
func getSession(ctx context.Context, sess *scs.SessionManager) *Session {
	return &Session{
		context: ctx,
		session: session,
	}
}

// Get returns the value for a given key from the session data. The return value
// has the type interface{} so will usually need to be type asserted
// before you can use it.
func (s *Session) Get(key string, pop ...bool) *sessionValue {

	var val interface{}

	if len(pop) > 0 && pop[0] {
		val = s.session.Pop(s.context, key)
	} else {
		val = s.session.Get(s.context, key)
	}

	return &sessionValue{
		value: val,
	}
}

// Set adds a key and corresponding value to the session data. Any existing
// value for the key will be replaced. The session data status will be set to
// Modified.
func (s *Session) Set(key string, value interface{}) {
	s.session.Put(s.context, key, value)
}

// SetMap adds a key and corresponding value to the session data from a Map.
// Any existing value for the key will be replaced. The session data
// status will be set to Modified.
func (s *Session) SetMap(m Map) {
	for key, value := range m {
		s.session.Put(s.context, key, value)
	}
}

// Destroy deletes the session data from the session store and sets the session
// status to Destroyed. Any further operations in the same request cycle will
// result in a new session being created.
func (s *Session) Destroy() error {
	return s.session.Destroy(s.context)
}

// Remove deletes the given key and corresponding value from the session data.
// The session data status will be set to Modified. If the key is not present
// this operation is a no-op.
func (s *Session) Remove(key string) {
	s.session.Remove(s.context, key)
}

// Clear removes all data for the current session. The session token and
// lifetime are unaffected. If there is no data in the current session this is
// a no-op.
func (s *Session) Clear() error {
	return s.session.Clear(s.context)
}

// Exists returns true if the given key is present in the session data.
func (s *Session) Exists(key string) bool {
	return s.session.Exists(s.context, key)
}

// Keys returns a slice of all key names present in the session data, sorted
// alphabetically. If the data contains no data then an empty slice will be
// returned.
func (s *Session) Keys() []string {
	return s.session.Keys(s.context)
}

// RenewToken updates the session data to have a new session token while
// retaining the current session data. The session lifetime is also reset and
// the session data status will be set to Modified.
//
// The old session token and accompanying data are deleted from the session store.
//
// To mitigate the risk of session fixation attacks, it's important that you call
// RenewToken before making any changes to privilege levels (e.g. login and
// logout operations). See https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md#renew-the-session-id-after-any-privilege-level-change
// for additional information.
func (s *Session) RenewToken() error {
	return s.session.RenewToken(s.context)
}

// MergeSession is used to merge in data from a different session in case strict
// session tokens are lost across an oauth or similar redirect flows. Use Clear()
// if no values of the new session are to be used.
func (s *Session) MergeSession(token string) error {
	return s.session.MergeSession(s.context, token)
}

// Status returns the current status of the session data.
func (s *Session) Status() scs.Status {
	return s.session.Status(s.context)
}

// RememberMe controls whether the session cookie is persistent (i.e  whether it
// is retained after a user closes their browser). RememberMe only has an effect
// if you have set SessionManager.Cookie.Persist = false (the default is true) and
// you are using the standard LoadAndSave() middleware.
func (s *Session) RememberMe(val bool) {
	s.session.RememberMe(s.context, val)
}

// Iterate retrieves all active (i.e. not expired) sessions from the store and
// executes the provided function fn for each session. If the session store
// being used does not support iteration then Iterate will panic.
func (s *Session) Iterate(fn func(session *Session) error) error {
	return s.session.Iterate(s.context, func(c context.Context) error {
		sess := getSession(c, s.session)
		return fn(sess)
	})
}

// Deadline returns the 'absolute' expiry time for the session. Please note
// that if you are using an idle timeout, it is possible that a session will
// expire due to non-use before the returned deadline.
func (s *Session) Deadline() time.Time {
	return s.session.Deadline(s.context)
}

// Token returns the session token. Please note that this will return the
// empty string "" if it is called before the session has been committed to
// the store.
func (s *Session) Token() string {
	return s.session.Token(s.context)
}

// SESSION VALUE

// Value returns the interface{} value for a given key from the session data.
func (s *sessionValue) Value() interface{} {
	return s.value
}

// String returns the string value for a given key from the session data.
// The zero value for a string ("") is returned if the key does not exist or the
// value could not be type asserted to a string.
func (s *sessionValue) String() string {
	str, ok := s.value.(string)
	if !ok {
		return ""
	}
	return str
}

// Bool returns the bool value for a given key from the session data. The
// zero value for a bool (false) is returned if the key does not exist or the
// value could not be type asserted to a bool.
func (s *sessionValue) Bool() bool {
	val, ok := s.value.(bool)
	if !ok {
		return false
	}
	return val
}

// Bytes returns the byte slice ([]byte) value for a given key from the session
// data. The zero value for a slice (nil) is returned if the key does not exist
// or could not be type asserted to []byte.
func (s *sessionValue) Bytes() []byte {
	val, ok := s.value.([]byte)
	if !ok {
		return nil
	}
	return val
}

// Time returns the time.Time value for a given key from the session data. The
// zero value for a time.Time object is returned if the key does not exist or the
// value could not be type asserted to a time.Time. This can be tested with the
// time.IsZero() method.
func (s *sessionValue) Time() time.Time {
	val, ok := s.value.(time.Time)
	if !ok {
		return time.Time{}
	}
	return val
}

// Int returns the int value for a given key from the session data. The
// zero value for an int (0) is returned if the key does not exist or the
// value could not be type asserted to an int.
func (s *sessionValue) Int() int {
	val, ok := s.value.(int)
	if !ok {
		return 0
	}
	return val
}

// Int32 returns the int32 value for a given key from the session data. The
// zero value for an int (0) is returned if the key does not exist or the
// value could not be type asserted to an int32.
func (s *sessionValue) Int32() int32 {
	val, ok := s.value.(int32)
	if !ok {
		return 0
	}
	return val
}

// Int64 returns the int64 value for a given key from the session data. The
// zero value for an int (0) is returned if the key does not exist or the
// value could not be type asserted to an int64.
func (s *sessionValue) Int64() int64 {
	val, ok := s.value.(int64)
	if !ok {
		return 0
	}
	return val
}

// Float32 returns the float32 value for a given key from the session data. The
// zero value for an float (0) is returned if the key does not exist or the
// value could not be type asserted to an float32.
func (s *sessionValue) Float32() float32 {
	val, ok := s.value.(float32)
	if !ok {
		return 0.0
	}
	return val
}

// Float64 returns the float64 value for a given key from the session data. The
// zero value for an float (0) is returned if the key does not exist or the
// value could not be type asserted to an float64.
func (s *sessionValue) Float64() float64 {
	val, ok := s.value.(float64)
	if !ok {
		return 0.0
	}
	return val
}
