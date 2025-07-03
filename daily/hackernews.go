package daily

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
)

var HNFRONTTEMPLATE = `
#### Hacker News
{{ range .Stories }}
- {{ .Title }}
  - Comments: {{ .Comments }}
  - ID: {{ .ID }}
{{ end }}
`

type HNFrontStory struct {
	Title    string
	Comments string
	ID       string
}

type HNFrontStories struct {
	Stories []HNFrontStory
}

func GetHackerNewsFront() (string, error) {
	req, err := http.NewRequest("GET", "https://news.ycombinator.com/front", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:140.0) Gecko/20100101 Firefox/140.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	stories := make([]HNFrontStory, 0, 10)
	doc.Find("tr.athing.submission").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i >= 10 {
			return false
		}
		// Get the story ID from the tr id attribute
		id, exists := s.Attr("id")
		if !exists {
			id = ""
		}

		// Get the title text
		title := strings.TrimSpace(s.Find("span.titleline").Text())

		// Find the next sibling tr, then the .subtext td, then the comments link
		subtext := s.Next().Find("td.subtext")
		comments := ""
		subtext.Find("a").Each(func(_ int, a *goquery.Selection) {
			href, _ := a.Attr("href")
			if strings.HasPrefix(href, "item?id=") && strings.Contains(a.Text(), "comment") {
				comments = strings.TrimSpace(strings.Replace(a.Text(), "\u00a0", " ", -1))
			}
		})

		stories = append(stories, HNFrontStory{
			Title:    title,
			Comments: comments,
			ID:       id,
		})
		return true
	})

	data := HNFrontStories{Stories: stories}
	tmpl, err := template.New("hnfront").Parse(HNFRONTTEMPLATE)
	if err != nil {
		return "", err
	}
	var ret strings.Builder
	if err := tmpl.Execute(&ret, data); err != nil {
		return "", err
	}
	return ret.String(), nil
}
