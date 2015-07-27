package api

import (
	"github.com/gocraft/web"
	"net/http"
	// "fmt"
)

func init() {
	router.Middleware((*Context).ValidateAwsCredentials)
	router.Get("/", (*Context).Root)
	router.Get("/vpc/:id/instances", (*Context).VpcListInstances)
	router.Delete("/instances/:id", (*Context).TerminateInstance)
	router.Put("/instances/:id/clone", (*Context).CloneInstance)
}

func (c *Context) Root(rw web.ResponseWriter, r *web.Request) {
	writeJson(rw, map[string]string{
		"status": "ok",
	})
}

func (c *Context) VpcListInstances(rw web.ResponseWriter, r *web.Request) {
	writeJson(rw, map[string]string{
		"status": "ok",
	})
}

func (c *Context) TerminateInstance(rw web.ResponseWriter, r *web.Request) {
	writeJson(rw, map[string]string{
		"status": "ok",
	})
}

func (c *Context) CloneInstance(rw web.ResponseWriter, r *web.Request) {
	writeJson(rw, map[string]string{
		"status": "ok",
	})
}

func (c *Context) ValidateAwsCredentials(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	if ok := validatePresenceRequest(r, "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"); !ok {
		rw.WriteHeader(http.StatusBadRequest)
		writeJson(rw, map[string]string{
			"error": "missing credentials",
		})
	} else {
		next(rw, r)
	}
}
