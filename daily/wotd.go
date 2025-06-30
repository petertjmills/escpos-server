package daily

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var WOTDTEMPLATE = `
#### Word of the day
{{.Word}} ({{.Pos}})
{{.Def}}
`

type Word struct {
	Word string
	Pos  string
	Def  string
}

// go to this webpage
// https://www.oed.com/?tl=true
// find div class="wotd"
// child h3 tag is the word
// child div class="wotdPos" is the type (verb, noun etc)
// child div class="wotdDef" is the definition

func GetWordOfTheDay() (string, error) {
	req, err := http.NewRequest("GET", "https://www.oed.com/?tl=true", nil)
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
	word := strings.TrimSpace(doc.Find(".wotd h3").Text())
	pos := strings.TrimSpace(doc.Find(".wotd .wotdPos").Text())
	def := strings.TrimSpace(doc.Find(".wotd .wotdDef").Text())

	w := Word{
		Word: word,
		Pos:  pos,
		Def:  def,
	}

	tmpl, err := template.New("weather").Parse(WOTDTEMPLATE)
	if err != nil {
		return "", err
	}

	var ret strings.Builder

	if err := tmpl.Execute(&ret, w); err != nil {
		return "", err
	}

	return ret.String(), nil

}
