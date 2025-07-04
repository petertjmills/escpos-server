package daily

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var POLLENCOUNTTEMPLATE = `
#### Pollen Count
|Type|Tonight|Tomorrow|{{ .DayAfter }}|
|---|---|---|---|
{{ range .PollenCounts }}
|{{ .Type }}|{{ .Tonight }}|{{ .Tomorrow }}|{{ .DayAfter }}|
{{ end }}
`

type PollenCount struct {
	Type     string
	Tonight  string
	Tomorrow string
	DayAfter string
}

type PollenData struct {
	PollenCounts []PollenCount
	DayAfter     string
}

func GetPollenCount() (string, error) {
	req, err := http.NewRequest("GET", "https://weather.com/en-GB/forecast/allergy/l/721f1cf300f0c0810eaf48183c120c433dda5c80efc87e1ff7fcb17cf705e1b2", nil)
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

	var pollenCounts []PollenCount
	dayAfter := ""

	// Find the pollen breakdown section
	doc.Find(".PollenBreakdown--breakdown--LjntG").Each(func(i int, s *goquery.Selection) {
		pollenType := strings.TrimSpace(s.Find(".PollenBreakdown--pollenType--vi9wa").Text())
		levels := s.Find(".PollenBreakdown--outlookLevels--RgaYH li")
		tonight := strings.TrimSpace(levels.Eq(0).Find("strong").Text())
		tomorrow := strings.TrimSpace(levels.Eq(1).Find("strong").Text())
		dayAfterVal := strings.TrimSpace(levels.Eq(2).Find("strong").Text())

		// Get the label for the third column (e.g., "Saturday")
		if dayAfter == "" {
			liText := levels.Eq(2).Text()
			parts := strings.Split(liText, ":")
			if len(parts) > 0 {
				dayAfter = strings.TrimSpace(parts[0])
			} else {
				dayAfter = "Day After"
			}
		}

		pollenCounts = append(pollenCounts, PollenCount{
			Type:     pollenType,
			Tonight:  tonight,
			Tomorrow: tomorrow,
			DayAfter: dayAfterVal,
		})
	})

	data := PollenData{
		PollenCounts: pollenCounts,
		DayAfter:     dayAfter,
	}

	tmpl, err := template.New("pollen").Parse(POLLENCOUNTTEMPLATE)
	if err != nil {
		return "", err
	}

	var ret strings.Builder
	if err := tmpl.Execute(&ret, data); err != nil {
		return "", err
	}

	return ret.String(), nil
}
