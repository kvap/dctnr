package mwapi

import (
	"encoding/json"
	"net/url"
	"net/http"
	"strconv"
	"fmt"
	"errors"
	"strings"
)

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

type Language struct {
	Langcode string
	Autonym  string
}

var languages *map[string]string

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
	if code != http.StatusOK {
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

func GetLanguages() (*map[string]string, error) {
	if languages == nil {
		ls := make(map[string]string)
		ls["en"] = "English"
		ls["ru"] = "Russian"
		languages = &ls
//		q := url.Values{}
//
//		q.Set("format", "json")
//		q.Set("formatversion", "2")
//
//		q.Set("action", "languagesearch")
//
//		resp, err := apiCall(&q, langcode)
//		apiCall(&q, "en")
	}
	return languages, nil
}

func GetFirstParagraph(title string, langcode string) ([]string, error) {
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
		return nil, err
	}

	result := make([]string, 0)
	for _, page := range resp.Query.Pages {
		fmt.Printf("page: %d: %s\n", page.Pageid, page.Title)

		absurl := fmt.Sprintf(
			"https://%s.wikipedia.org/wiki/%s",
			langcode, page.Title,
		)

		trimmed := strings.TrimSuffix(page.Extract, "…")
		if trimmed == "" {
			trimmed = page.Title
		}

		wrapped := fmt.Sprintf(
			"<a href=\"%s\">%s</a>",
			absurl, trimmed,
		)
		result = append(result, wrapped)
	}
	return result, nil
}

func SearchLanglinks(phrase, src, dst string, limit int) ([]string, error) {
	q := url.Values{}

	q.Set("format", "json")
	q.Set("formatversion", "2")
	q.Set("redirects", "")

	q.Set("action", "query")
	q.Set("generator", "search")
	q.Set("gsrsearch", phrase)
	q.Set("gsrlimit", strconv.Itoa(limit))
	q.Set("prop", "langlinks")
	q.Set("lllang", dst)

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
