package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"syscall"
	"github.com/alewmoose/show-lyrics/fetcher/azlyrics"
	"github.com/alewmoose/show-lyrics/songinfo"
)

func main() {
	home := os.Getenv("HOME")
	if home == "" {
		log.Fatal("HOME not found")
	}

	Songinfo, err := getSongInfo()
	if err != nil {
		log.Fatal(err)
	}

	artistP := replaceSlashes(Songinfo.Artist)
	titleP := replaceSlashes(Songinfo.Title)

	dotDir := path.Join(home, ".show-lyrics")
	cacheDir := path.Join(dotDir, "cache")
	cacheArtistDir := path.Join(cacheDir, artistP)
	songFile := path.Join(cacheArtistDir, titleP+".txt")

	_, err = os.Stat(songFile)
	if err == nil {
		err := execLess(songFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, dir := range []string{dotDir, cacheDir, cacheArtistDir} {
		err := mkdirUnlessExists(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	client := &http.Client{}

	lyrics, err := azlyrics.Fetch(client, Songinfo)
	if err != nil {
		log.Fatal(err)
	}

	lyrics = prepareLyrics(Songinfo, lyrics)

	err = ioutil.WriteFile(songFile, lyrics, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = execLess(songFile)
	if err != nil {
		log.Fatal(err)
	}
}

func prepareLyrics(si *songinfo.SongInfo, lyrics []byte) []byte {
	title := si.PrettyTitle()
	return []byte(title + "\n\n" + string(lyrics) + "\n")
}

func replaceSlashes(s string) string {
	return strings.Replace(s, "/", "_", -1)
}

func execLess(file string) error {
	lessBin, err := exec.LookPath("less")
	if err != nil {
		return err
	}
	err = syscall.Exec(lessBin, []string{"less", "-c", file}, os.Environ())
	if err != nil {
		return err
	}
	return nil
}

func getSongInfo() (*songinfo.SongInfo, error) {
	cmusStatus, err := getCmusStatus()
	if err != nil {
		return nil, err
	}
	Songinfo, err := parseCmusStatus(cmusStatus)
	if err != nil {
		return nil, err
	}
	return Songinfo, nil
}

func mkdirUnlessExists(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
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

func parseCmusStatus(cmusStatus []byte) (*songinfo.SongInfo, error) {
	artist := regexpMatch(artistRe, cmusStatus)
	title := regexpMatch(titleRe, cmusStatus)

	if artist == nil || title == nil {
		return nil, errors.New("Failed to parse cmus status")
	}

	si := songinfo.SongInfo{
		Artist: string(artist),
		Title:  string(title),
	}

	return &si, nil
}

