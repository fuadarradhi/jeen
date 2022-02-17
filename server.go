package jeen

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var database *sql.DB
var session *scs.SessionManager

type Router interface {
	chi.Router
}

var (
	_ Router = &chi.Mux{}
	_ Router = chi.Router(nil)
)

type Driver struct {
	Database *sql.DB
	Session  scs.Store
}
type Default struct {
	WithDatabase bool
	WithTimeout  time.Duration
	WithTemplate *Template
}

type Config struct {
	Driver  *Driver
	Default *Default
}

type Server struct {
	router       Router
	withDatabase bool
	withSession  bool
	withTimeout  time.Duration
	withTemplate *TemplateEngine
}

type HandlerServerFunc func(serv *Server)
type HandlerRouteFunc func(res *Resource)
type HandlerMiddlewareFunc func(res *Resource) bool

type Map map[string]interface{}

type Options func(s *Server)

func WithDatabase(usedb bool) Options {
	return func(s *Server) {
		s.withDatabase = usedb
	}
}

func WithTimeout(timeout time.Duration) Options {
	return func(s *Server) {
		s.withTimeout = timeout
	}
}

func WithTemplate(template *Template) Options {
	return func(s *Server) {
		s.withTemplate = NewTemplateEngine(
			mergeWithOldEngine(s.withTemplate.template, template),
		)
	}
}

func InitServer(cfg *Config) *Server {
	r := chi.NewRouter()

	defDb := false
	defSess := false
	defTimeout := 7 * time.Second
	var defTemplate *TemplateEngine

	if cfg.Default != nil {
		defDb = cfg.Default.WithDatabase
		defTimeout = cfg.Default.WithTimeout
		if cfg.Default.WithTemplate != nil {
			defTemplate = NewTemplateEngine(cfg.Default.WithTemplate)
		}

		if defTimeout < 2*time.Second {
			log.Fatal("Minimum timeout is 2 seconds.")
		}

	}

	if cfg.Driver != nil {
		if cfg.Driver.Database != nil {
			database = cfg.Driver.Database
		}
		if cfg.Driver.Session != nil {
			defSess = true
			session = scs.New()
			session.Store = cfg.Driver.Session
		}

		if defDb && cfg.Driver.Database == nil {
			log.Fatal("WithDatabase true, but driver not defined.")
		}
	}

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	return &Server{
		router:       r,
		withDatabase: defDb,
		withSession:  defSess,
		withTimeout:  defTimeout,
		withTemplate: defTemplate,
	}
}

func mergeWithOldEngine(old, new *Template) *Template {
	if new.Delims != nil {
		old.Delims = new.Delims
	}
	if new.Master != "" {
		old.Master = new.Master
	}
	if new.Root != "" {
		old.Root = new.Root
	}
	if new.Partials != nil {
		old.Partials = new.Partials
	}
	if new.Funcs != nil {
		old.Funcs = new.Funcs
	}
	return old
}

func (s *Server) busy(res *Resource, err error) {
	log.Println(err)
	// TODO: response busy
}

func (s *Server) newHandler(router Router, opts ...Options) *Server {
	serv := &Server{
		router:       router,
		withDatabase: s.withDatabase,
		withSession:  s.withSession,
		withTimeout:  s.withTimeout,
		withTemplate: s.withTemplate,
	}
	for _, opt := range opts {
		opt(serv)
	}
	return serv
}

func (s *Server) httpHandler(rw http.ResponseWriter, r *http.Request, handler interface{}, opts ...Options) bool {

	serv := &Server{
		withDatabase: s.withDatabase,
		withTimeout:  s.withTimeout,
		withSession:  s.withSession,
		withTemplate: s.withTemplate,
	}
	for _, opt := range opts {
		opt(serv)
	}

	// request with timeout context
	reqContext, cancel := context.WithTimeout(r.Context(), serv.withTimeout)
	defer cancel()
	r = r.WithContext(reqContext)

	res := createResource(rw, r, serv.withTemplate)

	if serv.withSession {
		res.Session = getSession(res.Context(), session)
	}

	if serv.withDatabase {
		db, err := conn(res.Context(), database)
		if err != nil {
			s.busy(res, err)
			return false
		}
		defer db.Close()

		res.Database = db
	}

	// Use goroutines to make sure
	// every request has a response when a timeout occurs
	//
	// [IMPORTANT]
	// cancel only applies to context,
	// for processes that don't have context
	// have to check in every process
	//
	// ... process 1 here
	//
	// select {
	//  case <-r.Context().Done():
	//  return
	//   defaults:
	// }
	//
	// ... process 2 here
	//
	processSuccess := make(chan bool)
	defer close(processSuccess)

	// the process is done in goroutine so that it can be canceled when
	// requesting timeout
	go func() {
		if h, ok := handler.(HandlerRouteFunc); ok {
			h(res)
		} else if h, ok := handler.(HandlerMiddlewareFunc); ok {
			if success := h(res); !success {
				processSuccess <- false
				return
			}
		} else {
			log.Fatal("Only HandlerRouteFunc and HandlerMiddlewareFunc are allowed")
		}
		processSuccess <- true
	}()

	select {

	// if request timeout show response busy.
	case <-reqContext.Done():
		s.busy(res, reqContext.Err())
		return false

	// if the process is successful, just return it.
	// response is done by main apps.
	case isSuccess := <-processSuccess:
		return isSuccess
	}
}

// Group creates a new inline-Mux with a fresh middleware stack. It's useful
// for a group of handlers along the same routing path that use an additional
// set of middlewares. See _examples/.
func (s *Server) Group(handler HandlerServerFunc, opts ...Options) *Server {
	s.router.Group(func(r chi.Router) {
		serv := s.newHandler(r, opts...)
		handler(serv)
	})
	return s
}

// Route creates a new Mux with a fresh middleware stack and mounts it
// along the `pattern` as a subrouter. Effectively, this is a short-hand
// call to Mount. See _examples/.
func (s *Server) Route(pattern string, handler HandlerServerFunc, opts ...Options) *Server {
	s.router.Route(pattern, func(r chi.Router) {
		serv := s.newHandler(r, opts...)
		handler(serv)
	})
	return s
}

// Mount attaches another http.Handler or jeen.Server as a subrouter along a routing
// path. It's very useful to split up a large API as many independent routers and
// compose them as a single service using Mount. See _examples/.
//
// Note that Mount() simply sets a wildcard along the `pattern` that will continue
// routing at the `handler`, which in most cases is another jeen.Server. As a result,
// if you define two Mount() routes on the exact same pattern the mount will panic.
func (s *Server) Mount(pattern string, handler HandlerServerFunc, opts ...Options) *Server {
	s.router.Mount(pattern, func() http.Handler {
		r := chi.NewRouter()
		serv := s.newHandler(r, opts...)
		handler(serv)
		return r
	}())
	return s
}

// Handle adds the route `pattern` that matches any http method to
// execute the `handler` jeen.HandlerServerFunc.
func (s *Server) Handle(pattern string, handler HandlerServerFunc, opts ...Options) *Server {
	s.router.Handle(pattern, func() http.Handler {
		r := chi.NewRouter()
		serv := s.newHandler(r, opts...)
		handler(serv)
		return r
	}())
	return s
}

// HandleFunc adds the route `pattern` that matches any http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) HandleFunc(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.HandleFunc(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Method adds the route `pattern` that matches `method` http method to
// execute the `handler` jeen.HandlerServerFunc.
func (s *Server) Method(method string, pattern string, handler HandlerServerFunc, opts ...Options) *Server {
	s.router.Method(method, pattern, func() http.Handler {
		r := chi.NewRouter()
		serv := s.newHandler(r, opts...)
		handler(serv)
		return r
	}())
	return s
}

// MethodFunc adds the route `pattern` that matches `method` http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) MethodFunc(method string, pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.MethodFunc(method, pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// NotFound sets a custom jeen.HandlerRouteFunc for routing paths that could
// not be found. The default 404 handler is `http.NotFound`.
func (s *Server) NotFound(handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.NotFound(func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// MethodNotAllowed sets a custom jeen.HandlerRouteFunc for routing paths where the
// method is unresolved. The default handler returns a 405 with an empty body.
func (s *Server) MethodNotAllowed(handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.MethodNotAllowed(func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Connect adds the route `pattern` that matches a CONNECT http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Connect(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Connect(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Delete adds the route `pattern` that matches a DELETE http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Delete(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Delete(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Get adds the route `pattern` that matches a GET http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Get(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Get(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Head adds the route `pattern` that matches a HEAD http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Head(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Head(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Options adds the route `pattern` that matches a OPTIONS http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Options(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Options(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Patch adds the route `pattern` that matches a PATCH http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Patch(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Patch(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Post adds the route `pattern` that matches a POST http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Post(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Post(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Put adds the route `pattern` that matches a PUT http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Put(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Put(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Trace adds the route `pattern` that matches a TRACE http method to
// execute the `handler` jeen.HandlerRouteFunc.
func (s *Server) Trace(pattern string, handler HandlerRouteFunc, opts ...Options) *Server {
	s.router.Trace(pattern, func(rw http.ResponseWriter, r *http.Request) {
		s.httpHandler(rw, r, handler, opts...)
	})
	return s
}

// Use appends a middleware handler to the Mux middleware stack.
//
// The middleware stack for any Mux will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next http.Handler.
func (s *Server) Use(handler HandlerMiddlewareFunc, opts ...Options) {
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			success := s.httpHandler(rw, r, handler, opts...)
			if !success {
				return
			}
			next.ServeHTTP(rw, r)
		})
	})
}

// With adds inline middlewares for an endpoint handler.
func (s *Server) With(handler HandlerMiddlewareFunc, opts ...Options) {
	s.router.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			success := s.httpHandler(rw, r, handler, opts...)
			if !success {
				return
			}
			next.ServeHTTP(rw, r)
		})
	})
}

// Close server and all resource
func (s *Server) Close() {
	if database != nil {
		database.Close()
	}
	log.Println("Thank you, server has been stopped.")
}

// ListenAndServe listens on the TCP network address addr and then calls Serve
// with handler to handle requests on incoming connections. Accepted connections
// are configured to enable TCP keep-alives.
func (s *Server) ListenAndServe(addr string) {

	// use session only if declared
	var handler http.Handler
	if s.withSession {
		handler = session.LoadAndSave(s.router)
	} else {
		handler = s.router
	}

	// router
	server := &http.Server{
		Addr:    addr,
		Handler: handler,

		// http default timeout is 5 minutes
		// for request set timeout in context
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}

	// Close cannot be called only in defer.
	// if terminate with ctrl+c, defer is not called
	// so need to do it with goroutine to check notify
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals to terminate or quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		fmt.Println("")
		log.Println("Please wait...")

		// Shutdown will wait for all contexts to finish for up to 10 seconds.
		// If after 30 seconds it has not finished, it will be force stopped.
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 10*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}

		serverStopCtx()
	}()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-serverCtx.Done()
}
