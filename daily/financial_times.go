package daily

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

var NEWSTEMPLATE = `
#### News
{{ range .Headlines }}
- {{ . }}
{{ end }}
`

type News struct {
	Headlines []string
}

// go to this webpage
// https://www.oed.com/?tl=true
// find div class="wotd"
// child h3 tag is the word
// child div class="wotdPos" is the type (verb, noun etc)
// child div class="wotdDef" is the definition

func GetNews() (string, error) {
	req, err := http.NewRequest("GET", "https://www.ft.com/", nil)
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

	w := News{
		Headlines: make([]string, 0),
	}

	doc.Find(".headline").Each(func(i int, s *goquery.Selection) {
		headline := strings.TrimSpace(s.Text())
		if len(w.Headlines) >= 10 {
			return
		}
		text := strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, headline)

		w.Headlines = append(w.Headlines, text)

	})

	tmpl, err := template.New("news").Parse(NEWSTEMPLATE)
	if err != nil {
		return "", err
	}

	var ret strings.Builder

	if err := tmpl.Execute(&ret, w); err != nil {
		return "", err
	}

	return ret.String(), nil

}
