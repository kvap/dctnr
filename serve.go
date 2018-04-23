package main

import (
	"fmt"
	"strings"
	"errors"
	"net/http"
	"time"
	"encoding/json"

	"github.com/gocraft/web"

	"github.com/kvap/dctnr/db"
	"github.com/kvap/dctnr/mwapi"
)

type Context struct {
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
	maxage := 24 * time.Hour

	if err := checkSearchRequest(phrase, src, dst); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	response, cached := db.GetCached(phrase, src, dst, maxage)
	db.LogAccess(req.RemoteAddr, phrase, src, dst, cached)
	if !cached {
		limit := 5
		titles, err := mwapi.SearchLanglinks(phrase, src, dst, limit)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}

		paragraphs, err := mwapi.GetFirstParagraphs(titles, dst)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}

		response = strings.Join(paragraphs, "")
		db.PutCached(phrase, src, dst, response)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(response))
}

func (c *Context) Stats(rw web.ResponseWriter, req *web.Request) {
	stats, err := db.GetStats()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	body, err := json.MarshalIndent(stats, "", "    ")
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func (c *Context) Main(rw web.ResponseWriter, req *web.Request) {
	http.Redirect(rw, req.Request, "/index.html", http.StatusFound)
}

func main() {
	db.Init()

	router := web.New(Context{})

	router.Middleware(web.LoggerMiddleware)
	router.Middleware(web.ShowErrorsMiddleware)
	router.Middleware(web.StaticMiddleware("static"))

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
