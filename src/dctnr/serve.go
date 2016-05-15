package main

import (
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"bytes"
	"errors"
//	"time"
	"net/http"
	"net/url"
	"encoding/json"
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

type ApiLangLink struct {
	Lang  string `json:"lang"`
	Title string `json:"title"`
}

type ApiPage struct {
	Pageid    int           `json:"pageid"`
	Title     string        `json:"title"`
	Langlinks []ApiLangLink `json:"langlinks"`
	Extract   string        `json:"extract"`
}

type ApiFromTo struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ApiResponse struct {
	Complete bool `json:"batchcomplete"`
	Query struct {
		Normalized []ApiFromTo `json:"normalized"`
		Redirects  []ApiFromTo `json:"redirects"`
		Pages      []ApiPage   `json:"pages"`
	} `json:"query"`
}

func getElementById(node *html.Node, id string) *html.Node {
	if node.Type == html.ElementNode {
		for _, a := range node.Attr {
			if a.Key == "id" && a.Val == id {
				return node
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if found := getElementById(c, id); found != nil {
			return found
		}
	}
	return nil
}

func getChildByTag(node *html.Node, tag string) *html.Node {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "p" {
			return c
		}
	}
	return nil
}

func renderNode(node *html.Node) (string, error) {
	buf := new(bytes.Buffer)
	if err := html.Render(buf, node); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func apiCall(q *url.Values, langcode string) (*ApiResponse, error) {
	absurl := fmt.Sprintf(
		"https://%s.wikipedia.org/w/api.php?%s",
		langcode, q.Encode(),
	)

	fmt.Println("getting " + absurl)

	res, err := http.Get(absurl)
	if err != nil {
		return nil, err
	}

	code := res.StatusCode
	if code != 200 {
		return nil, errors.New("'not OK' response from wikipedia")
	}

	dec := json.NewDecoder(res.Body)
	var resp ApiResponse
	err = dec.Decode(&resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func getFirstParagraph(title string, langcode string) (string, error) {
	q := url.Values{}

	q.Set("format", "json")
	q.Set("formatversion", "2")
	q.Set("redirects", "")

	q.Set("action", "query")
	q.Set("prop", "extracts")
	q.Set("exintro", "")
	q.Set("exchars", "1024")
	q.Set("titles", title)

	resp, err := apiCall(&q, langcode)
	if err != nil {
		return "", err
	}

	for _, page := range resp.Query.Pages {
		return page.Extract, nil
	}

	return "", errors.New("no page found")
}

func translateTitle(title string, src string, dst string) ([]string, error) {
	q := url.Values{}

	q.Set("format", "json")
	q.Set("formatversion", "2")
	q.Set("redirects", "")

	q.Set("action", "query")
	q.Set("prop", "langlinks")
	q.Set("lllang", dst)
	q.Set("titles", title)

	resp, err := apiCall(&q, src)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, page := range resp.Query.Pages {
		for _, link := range page.Langlinks {
			result = append(result, link.Title)
		}
	}

	return result, nil
}

func (c *Context) Search(rw web.ResponseWriter, req *web.Request) {
	q := req.URL.Query()
	srctitle := q.Get("phrase")
	src := q.Get("src")
	dst := q.Get("dst")

	dsttitles, err := translateTitle(srctitle, src, dst)
	if err != nil {
		rw.WriteHeader(500)
		return
	}

	var reply string
	for _, t := range dsttitles {
		p, err := getFirstParagraph(t, dst)
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
