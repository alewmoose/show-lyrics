package main

import (
	"fmt"
	"os/exec"
	"log"
	"regexp"
	"errors"
	"bytes"
	"net/http"
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

	lyrics := fetchLyrics(client, songinfo)
	fmt.Println(lyrics)
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

	for _, str := range [][]byte{artist,title} {
		str = bytes.ToLower(str)
		str = weirdRe.ReplaceAll(str, []byte{})
	}

	fmt.Println(string(artist))
	fmt.Println(string(title))

	return ""
}

func fetchLyrics (client *http.Client, si *songInfo) []byte {
	url := makeURL(si)
	_ = url

	return []byte{}
}
