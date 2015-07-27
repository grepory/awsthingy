package api

import (
	"github.com/gocraft/health"
	"github.com/gocraft/web"
	"github.com/nu7hatch/gouuid"
	"fmt"
	"net/http"
	"errors"
	"runtime"
	"io"
	"strconv"
	"encoding/json"
)

type Context struct {
	Job   *health.Job
	Panic bool
}

var (
	stream = health.NewStream()
	router = web.New(Context{})
)

func init() {
	router.Middleware((*Context).Log)
	router.Middleware((*Context).CatchPanics)
	router.Middleware((*Context).SetContentType)
	router.Middleware((*Context).Cors)
	router.NotFound((*Context).NotFound)
}

func ListenAndServe(addr string, sink io.Writer) {
	if sink != nil {
		stream.AddSink(&health.WriterSink{sink})
	}
	stream.Event("api.listen-and-serve")
	http.ListenAndServe(addr, router)
}

func (c *Context) Log(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	c.Job = stream.NewJob(r.RoutePath())

	id, err := uuid.NewV4()
	if err == nil {
		c.Job.KeyValue("request-id", id.String())
	}

	path := r.URL.Path
	c.Job.EventKv("api.request", health.Kvs{"path": path})

	next(rw, r)

	code := rw.StatusCode()
	kvs := health.Kvs{
		"code": fmt.Sprint(code),
		"path": path,
	}

	// Map HTTP status code to category.
	var status health.CompletionStatus
	if c.Panic {
		status = health.Panic
	} else if code < 400 {
		status = health.Success
	} else if code == 422 {
		status = health.ValidationError
	} else if code < 500 {
		status = health.Junk // 404, 401
	} else {
		status = health.Error
	}
	c.Job.CompleteKv(status, kvs)
}

func (c *Context) CatchPanics(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	defer func() {
		if err := recover(); err != nil {
			c.Panic = true

			const size = 4096
			stack := make([]byte, size)
			stack = stack[:runtime.Stack(stack, false)]

			// err turns out to be interface{}, of actual type "runtime.errorCString"
			// The health package kinda wants an error. Luckily, the err sprints nicely via fmt.
			errorishError := errors.New(fmt.Sprint(err))

			c.Job.EventErrKv("panic", errorishError, health.Kvs{"stack": string(stack)})
			renderServerError(rw)
		}
	}()

	next(rw, r)
}

func (c *Context) SetContentType(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	header := rw.Header()
	header.Set("Content-Type", "application/json; charset=utf-8")
	next(rw, r)
}

func (c *Context) Cors(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	header := rw.Header()
	header.Set("Access-Control-Allow-Origin", "*")
	next(rw, r)
}

func (c *Context) NotFound(rw web.ResponseWriter, r *web.Request) {
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rw, "{\"errors\":{\"error\":\"not_found\"}}\n")
}

func renderServerError(rw web.ResponseWriter) {
	rw.WriteHeader(500)
	fmt.Fprintf(rw, "\"not good\"\n")
}

func validatePresencePathParams(params map[string]string, keys... string) bool {
	ok := true
	for _, k := range keys {
		if params[k] == "" {
			ok = false
			break
		}
	}
	return ok
}

// FIXME: combine presence validation of path params and form value
// FIXME: this is case sensitive which is wrong
func validatePresenceRequest(r *web.Request, keys... string) bool {
	ok := true
	for _, k := range keys {
		if r.FormValue(k) == "" {
			ok = false
			break
		}
	}
	return ok
}

func formvalueInt(value string, thedefault int) (int, error) {
	if value == "" {
		return thedefault, nil
	} else {
		i, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		} else {
			return i, nil
		}
	}
}

func writeJson(rw web.ResponseWriter, data interface{}) {
	encoder := json.NewEncoder(rw)
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
}
