package main

import (
	"fmt"
	"strings"
	"errors"
	"net/http"
	"github.com/gocraft/web"
	"dctnr/db"
	"dctnr/mwapi"
)

type Context struct {
	HelloCount int
}

func (c *Context) PingDB(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	err := db.Ping()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("db unavailable: " + err.Error()))
		return
	}
	next(rw, req)
}

func checkSearchRequest(phrase, src, dst string) error {
	if len(phrase) == 0 {
		return errors.New("phrase is empty")
	}

	languages, _ := mwapi.GetLanguages()

	if _, found := (*languages)[src]; !found {
		return errors.New(fmt.Sprintf("unknown src language '%s'", src))
	}

	if _, found := (*languages)[dst]; !found {
		return errors.New(fmt.Sprintf("unknown dst language '%s'", dst))
	}

	return nil
}

func (c *Context) Search(rw web.ResponseWriter, req *web.Request) {
	q := req.URL.Query()
	phrase := q.Get("phrase")
	src := q.Get("src")
	dst := q.Get("dst")

	if err := checkSearchRequest(phrase, src, dst); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	limit := 5
	titles, err := mwapi.SearchLanglinks(phrase, src, dst, limit)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	paragraphs := make([]string, 0)
	for _, t := range titles {
		newpars, err := mwapi.GetFirstParagraph(t, dst)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		paragraphs = append(paragraphs, newpars...)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(strings.Join(paragraphs, "")))
}

func (c *Context) Stats(rw web.ResponseWriter, req *web.Request) {
	fmt.Println(strings.Repeat("Stats ", c.HelloCount), "World! ")
}

func (c *Context) Main(rw web.ResponseWriter, req *web.Request) {
	fmt.Println(strings.Repeat("Main ", c.HelloCount), "World! ")
}

func main() {
	db.Init()

	router := web.New(Context{})

	router.Middleware(web.LoggerMiddleware)
	router.Middleware(web.ShowErrorsMiddleware)
	router.Middleware(web.StaticMiddleware("static"))
	router.Middleware((*Context).PingDB)

	router.Get("/search", (*Context).Search)
	router.Get("/stats", (*Context).Stats)
	router.Get("/", (*Context).Main)

	fmt.Println("starting the listener")
	err := http.ListenAndServe("0.0.0.0:4000", router)
	if err != nil {
		fmt.Println(err)
	}

	db.Finalize()
}
