package azlyrics

import (
	"bytes"
	"errors"
	"github.com/alewmoose/show-lyrics/songinfo"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"net/http"
	"regexp"
)

func Fetch(client *http.Client, si *songinfo.SongInfo) ([]byte, error) {
	reqUrl := makeURL(si)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != 200 {
		return []byte{}, errors.New(resp.Status)
	}

	utf8, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	defer resp.Body.Close()
	if err != nil {
		return []byte{}, err
	}

	body, err := ioutil.ReadAll(utf8)
	if err != nil {
		return []byte{}, err
	}

	lyrics, err := parseLyrics(body)
	if err != nil {
		return []byte{}, err
	}

	return lyrics, nil
}

func makeURL(si *songinfo.SongInfo) string {
	artist := []byte(si.Artist)
	title := []byte(si.Title)

	theRe := regexp.MustCompile(`(?i)^the `)
	weirdRe := regexp.MustCompile(`(?i)[^a-z0-9]`)

	artist = theRe.ReplaceAll(artist, []byte{})

	artist = bytes.ToLower(artist)
	title = bytes.ToLower(title)

	for _, str := range []*[]byte{&artist, &title} {
		*str = bytes.ToLower(*str)
		*str = weirdRe.ReplaceAll(*str, []byte{})
	}

	url := "https://www.azlyrics.com/lyrics/"
	url += string(artist) + "/" + string(title) + ".html"

	return url
}

func htmlStrip(html []byte) []byte {
	commentsRe := regexp.MustCompile(`(?s)<!--.*?-->`)
	brRe := regexp.MustCompile(`<br/?>`)
	tagsRe := regexp.MustCompile(`<[^<>]+>`)

	html = commentsRe.ReplaceAll(html, []byte{})
	html = brRe.ReplaceAll(html, []byte{})
	html = tagsRe.ReplaceAll(html, []byte{})
	html = bytes.TrimSpace(html)

	return html
}

func parseLyrics(lyricsHtml []byte) ([]byte, error) {
	re := regexp.MustCompile(
		`(?s)<div[^<>]*?class="lyricsh"[^<>]*?>.*?</div>\s*?` +
			`<div[^<>]*?>.*?</div>\s*` +
			`.*?` +
			`<div[^<>]*?>(.*?)</div>`)

	match := re.FindAllSubmatch(lyricsHtml, 1)
	if match == nil {
		return []byte{}, errors.New("Failed to parse html")
	}

	lyrics := htmlStrip(match[0][1])
	return lyrics, nil
}
