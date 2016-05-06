package main

import (
	"time"
	"fmt"
	"strings"
	"net/http"
	"github.com/gocraft/web"
	"dctnr/db"
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

func (c *Context) SetHelloCount(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	time.Sleep(1000 * time.Millisecond)
	c.HelloCount = 3
	next(rw, req)
}

func (c *Context) SayHello(rw web.ResponseWriter, req *web.Request) {
	fmt.Print(strings.Repeat("Hello ", c.HelloCount), "World! ")
}

func (c *Context) Search(rw web.ResponseWriter, req *web.Request) {
	phrase := req.URL.Query().Get("phrase");
	rw.WriteHeader(http.StatusOK);
	rw.Write([]byte("your phrase was: '" + phrase + "'"));
}

func (c *Context) Stats(rw web.ResponseWriter, req *web.Request) {
	fmt.Print(strings.Repeat("Stats ", c.HelloCount), "World! ")
}

func (c *Context) Main(rw web.ResponseWriter, req *web.Request) {
	fmt.Print(strings.Repeat("Main ", c.HelloCount), "World! ")
}

func main() {
	for {
		err := db.Init()
		if err == nil {
			break
		}
		fmt.Println(err)
	}

	router := web.New(Context{})

	router.Middleware(web.LoggerMiddleware)
	router.Middleware(web.ShowErrorsMiddleware)
	router.Middleware(web.StaticMiddleware("static"))
	router.Middleware((*Context).PingDB)
	router.Middleware((*Context).SetHelloCount)

	router.Get("/search", (*Context).Search)
	router.Get("/stats", (*Context).Stats)
	router.Get("/", (*Context).Main)

	router.Get("/hello", (*Context).SayHello)

	fmt.Println("starting the listener")
	err := http.ListenAndServe("0.0.0.0:4000", router)
	if err != nil {
		fmt.Println(err)
	}

	db.Finalize()
}
