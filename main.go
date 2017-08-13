package main

import (
	"fmt"
	"os/exec"
	"log"
	"regexp"
	"errors"
	"bytes"
	"net/http"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	// "syscall"
)

type songInfo struct {
	artist, title string
}

func main() {
	// TODO
	// exit status
	// strings vs bytes slices
	// http client : what are default settings?

	cmusStatus, err := getCmusStatus()
	if err != nil {
		log.Fatal(err)
	}

	songinfo, err := parseCmusStatus(cmusStatus)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	fmt.Println(songinfo)

	lyrics, err := fetchLyrics(client, songinfo)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(lyrics))
}

func getCmusStatus() ([]byte, error) {
	cmd := exec.Command("cmus-remote", "-Q")
	return cmd.Output()
}

var artistRe = regexp.MustCompile(`(?m)^tag\s+artist\s+(.+)\s*$`)
var titleRe = regexp.MustCompile(`(?m)^tag\s+title\s+(.+)\s*$`)

func regexpMatch(re *regexp.Regexp, buf []byte) []byte {
	match := re.FindAllSubmatch(buf, 1)
	if len(match) > 0 {
		return match[0][1]
	}
	return nil
}

func parseCmusStatus(cmusStatus []byte) (*songInfo, error) {
	artist := regexpMatch(artistRe, cmusStatus)
	title := regexpMatch(titleRe, cmusStatus)

	if artist == nil || title == nil {
		return nil, errors.New("Failed to parse cmus status")
	}

	si := songInfo {
		artist: string(artist),
		title: string(title),
	}

	return &si, nil
}

func makeURL(si *songInfo) string {
	artist := []byte(si.artist)
	title := []byte(si.title)

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
	url += string(artist) + "/" + string(title) + ".html";

	return url
}

func fetchLyrics (client *http.Client, si *songInfo) ([]byte, error) {
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

func htmlStrip(html []byte) []byte {
	commentsRe := regexp.MustCompile(`(?s)<!--.*?-->`)
	brRe := regexp.MustCompile(`<br/?>`)

	html = commentsRe.ReplaceAll(html, []byte{})
	html = brRe.ReplaceAll(html, []byte{})

	// TODO
	// trim leading|trainling whitespace for every line

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

