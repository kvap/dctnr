package main

import (
	"time"
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"bytes"
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

func (c *Context) SetHelloCount(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	time.Sleep(1000 * time.Millisecond)
	c.HelloCount = 3
	next(rw, req)
}

func (c *Context) SayHello(rw web.ResponseWriter, req *web.Request) {
	fmt.Println(strings.Repeat("Hello ", c.HelloCount), "World! ")
}

func deepGet(obj interface{}, path ...interface{}) (interface{}, bool) {
	var val interface{} = obj

	for _, keystr := range path {
		if key, ok := keystr.(int); ok {
			if arr, ok := val.([]interface{}); ok {
				val = arr[key];
			} else {
				return val, false;
			}
		} else if key, ok := keystr.(string); ok {
			if m, ok := val.(map[string]interface{}); ok {
				val = m[key];
			} else {
				return val, false;
			}
		}
	}

	return val, true
}

type LL struct {
	Lang      string `json:"lang"`
	URL       string `json:"url"`
	Langname  string `json:"langname"`
	Autonym   string `json:"autonym"`
	Autotitle string `json:"*"`
}

type LLPage struct {
	Pageid    int    `json:"pageid"`
	Title     string `json:"title"`
	Langlinks []LL   `json:"langlinks"`
}

type LLResponse struct {
	Query struct {
		Normalized struct {
			From string            `json:"from"`
			To   string            `json:"to"`
		} `json:"normalized"`
		Pages        map[string]LLPage `json:"pages"`
	} `json:"query"`
}

//func fetchArticle(url string) (io.Reader, error) {
//	res, err := http.Get(url)
//	if err != nil {
//		return nil, err
//	}
//	return res.Body, nil
//}

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

func getFirstParagraph(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	root, err := html.Parse(res.Body)
	if err != nil {
		return "", err
	}

	if c := getElementById(root, "mw-content-text"); c != nil {
		if p := getChildByTag(c, "p"); p != nil {
			return renderNode(p)
		}
	}

	return "", nil
}

func fetchLanguages(phrase string, src string, dst string) (string, int) {
	q := url.Values{}
	q.Set("action", "query")
	q.Set("prop", "langlinks")
	q.Set("format", "json")
	q.Set("lllang", dst)
	q.Set("llprop", "url|langname|autonym")
	q.Set("titles", phrase)

	url := fmt.Sprintf(
		"https://%s.wikipedia.org/w/api.php?%s",
		src, q.Encode(),
	)

	fmt.Println("getting " + url)

	res, err := http.Get(url)
	if err != nil {
		return "something went wrong", 500
	}

	code := res.StatusCode

	if code != 200 {
		return "wrong return code during article fetching", code
	}

	dec := json.NewDecoder(res.Body)
	var resp LLResponse
	err = dec.Decode(&resp);
	if err != nil {
		return "something went wrong during response decoding", 500
	}

	var result string

	for _, page := range resp.Query.Pages {
		for _, link := range page.Langlinks {
			p, err := getFirstParagraph(link.URL)
			if err != nil {
				result += "\nerror"
			}
			result += "\n" + p
		}
	}

	return result, code
}

func (c *Context) Search(rw web.ResponseWriter, req *web.Request) {
	q := req.URL.Query()
	phrase := q.Get("phrase")
	src := q.Get("src")
	dst := q.Get("dst")
	reply, code := fetchLanguages(phrase, src, dst)
	rw.WriteHeader(code)
	rw.Write([]byte(reply))
}

func (c *Context) Stats(rw web.ResponseWriter, req *web.Request) {
	fmt.Println(strings.Repeat("Stats ", c.HelloCount), "World! ")
}

func (c *Context) Main(rw web.ResponseWriter, req *web.Request) {
	fmt.Println(strings.Repeat("Main ", c.HelloCount), "World! ")
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
