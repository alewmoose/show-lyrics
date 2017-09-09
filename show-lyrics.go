package main

import (
	"errors"
	"github.com/alewmoose/show-lyrics/cache"
	"github.com/alewmoose/show-lyrics/fetcher/azlyrics"
	"github.com/alewmoose/show-lyrics/player/cmus"
	"github.com/alewmoose/show-lyrics/player/mocp"
	"github.com/alewmoose/show-lyrics/songinfo"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
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

	dotDir := path.Join(home, ".show-lyrics")
	cacheDir := path.Join(dotDir, "cache")

	for _, dir := range []string{dotDir, cacheDir} {
		err := mkdirUnlessExists(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	lyricsCache := cache.New(cacheDir, songinfo)

	if lyricsCache.Exists() {
		err := execLess(lyricsCache.FilePath())
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

	err = lyricsCache.Store(lyrics)
	if err != nil {
		log.Fatal(err)
	}

	err = execLess(lyricsCache.FilePath())
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
