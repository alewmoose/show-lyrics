package main

import (
	"errors"
	"github.com/alewmoose/show-lyrics/fetcher/azlyrics"
	"github.com/alewmoose/show-lyrics/player/cmus"
	"github.com/alewmoose/show-lyrics/player/mocp"
	"github.com/alewmoose/show-lyrics/songinfo"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"syscall"
)

func main() {
	home := os.Getenv("HOME")
	if home == "" {
		log.Fatal("HOME not found")
	}

	songinfo, err := getSongInfo()
	if err != nil {
		log.Fatal(err)
	}

	artistP := replaceSlashes(songinfo.Artist)
	titleP := replaceSlashes(songinfo.Title)

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

	lyrics, err := fetchLyrics(client, songinfo)
	if err != nil {
		log.Fatal(err)
	}

	lyrics = prepareLyrics(songinfo, lyrics)

	err = ioutil.WriteFile(songFile, lyrics, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = execLess(songFile)
	if err != nil {
		log.Fatal(err)
	}
}

var parensRe = regexp.MustCompile(`\(.+\)$`)

func fetchLyrics(c *http.Client, si *songinfo.SongInfo) ([]byte, error) {
	lyrics, err := azlyrics.Fetch(c, si)
	if err == nil {
		return lyrics, err
	}
	if err.Error() != "404 Not Found" {
		return lyrics, err
	}
	if parensRe.MatchString(si.Title) == false {
		return lyrics, err
	}
	title := parensRe.ReplaceAllString(si.Title, "")
	if len(title) == 0 {
		return lyrics, err
	}
	newSi := songinfo.SongInfo{Artist: si.Artist, Title: title}
	return fetchLyrics(c, &newSi)
}

func getSongInfo() (*songinfo.SongInfo, error) {
	type songInfoResult struct {
		songinfo *songinfo.SongInfo
		err      error
	}

	cmusSi, cmusErr := cmus.GetSongInfo()
	mocpSi, mocpErr := mocp.GetSongInfo()

	if cmusErr != nil && mocpErr != nil {
		return nil, errors.New("No players running")
	}
	if mocpErr != nil {
		return cmusSi, nil
	}
	return mocpSi, nil
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
