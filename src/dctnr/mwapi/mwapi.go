package mwapi

import (
	"encoding/json"
	"net/url"
	"net/http"
	"fmt"
	"errors"
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

func GetFirstParagraph(title string, langcode string) (string, error) {
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

func TranslateTitle(title string, src string, dst string) ([]string, error) {
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

