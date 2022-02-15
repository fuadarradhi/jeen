package jeen

import (
	"context"

	"github.com/alexedwards/scs/v2"
)

type Session struct {
	session *scs.SessionManager
	context context.Context
}

func getSession(ctx context.Context, sess *scs.SessionManager) *Session {
	return &Session{
		context: ctx,
		session: session,
	}
}
