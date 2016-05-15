package main

import (
	"fmt"
	"strings"
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

func (c *Context) Search(rw web.ResponseWriter, req *web.Request) {
	q := req.URL.Query()
	srctitle := q.Get("phrase")
	src := q.Get("src")
	dst := q.Get("dst")

	dsttitles, err := mwapi.TranslateTitle(srctitle, src, dst)
	if err != nil {
		rw.WriteHeader(500)
		return
	}

	var reply string
	for _, t := range dsttitles {
		p, err := mwapi.GetFirstParagraph(t, dst)
		if err != nil {
			reply += err.Error()
		} else {
			reply += p
		}
	}

	rw.WriteHeader(200)
	rw.Write([]byte(reply))
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
