package lyrics_wikia

import (
	"bytes"
	"errors"
	"github.com/alewmoose/show-lyrics/songinfo"
	"golang.org/x/net/html/charset"
	"html"
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

var spaceRe = regexp.MustCompile(`\s+`)

func makeURL(si *songinfo.SongInfo) string {
	artist := []byte(si.Artist)
	title := []byte(si.Title)
	for _, str := range []*[]byte{&artist, &title} {
		*str = spaceRe.ReplaceAll(*str, []byte{byte('_')})
	}

	url := "http://lyrics.wikia.com/wiki/"
	url += string(artist) + ":" + string(title)

	return url
}

var commentsRe = regexp.MustCompile(`(?s)<!--.*?-->`)
var brRe = regexp.MustCompile(`<br\s*/?>`)
var tagsRe = regexp.MustCompile(`<[^<>]+>`)

func htmlStrip(h []byte) []byte {
	h = commentsRe.ReplaceAll(h, []byte{})
	h = brRe.ReplaceAll(h, []byte{byte('\n')})
	h = tagsRe.ReplaceAll(h, []byte{})
	h = bytes.TrimSpace(h)
	h = []byte(html.UnescapeString(string(h)))

	return h
}

var parseLyricsRe = regexp.MustCompile(
	`(?s)<div[^<>]*class='lyricbox'[^<>]*>` +
	`(.*?)` +
	`<div class='lyricsbreak'></div>\s*</div>`)

func parseLyrics(lyricsHtml []byte) ([]byte, error) {
	match := parseLyricsRe.FindAllSubmatch(lyricsHtml, 1)
	if match == nil {
		return []byte{}, errors.New("Failed to parse html")
	}

	lyrics := htmlStrip(match[0][1])
	return lyrics, nil
}
