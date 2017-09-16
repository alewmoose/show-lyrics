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
	"time"
)

func main() {
	home := os.Getenv("HOME")
	if home == "" {
		log.Fatal("HOME not found")
	}

	dotDir := path.Join(home, ".show-lyrics")
	cacheDir := path.Join(dotDir, "cache")

	for _, dir := range []string{dotDir, cacheDir} {
		err := mkdirUnlessExists(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	client := &http.Client{}

	err := mainLoop(client, cacheDir)
	if err != nil {
		log.Fatal(err)
	}
}

func mainLoop(client *http.Client, cacheDir string) error {
	var prevSongInfo songinfo.SongInfo
	var cmd *exec.Cmd
	for ; ; time.Sleep(5 * time.Second) {
		songinfo, err := getSongInfo()
		if err != nil {
			return err
		}
		if *songinfo == prevSongInfo {
			continue
		}
		prevSongInfo = *songinfo

		lyricsCache := cache.New(cacheDir, songinfo)
		if !lyricsCache.Exists() {
			lyrics, err := fetchLyrics(client, songinfo)
			if err != nil {
				return err
			}
			lyrics = prepareLyrics(songinfo, lyrics)
			err = lyricsCache.Store(lyrics)
			if err != nil {
				return err
			}
		}

		if cmd != nil {
			sigErr := cmd.Process.Signal(syscall.SIGTERM)
			if sigErr != nil {
				return sigErr
			}
			processState, waitErr := cmd.Process.Wait()
			if processState.Exited() {
				return nil
			}
			if waitErr != nil {
				return waitErr
			}
		}
		cmd = exec.Command("less", "-c", lyricsCache.FilePath())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		startErr := cmd.Start()
		if startErr != nil {
			return startErr
		}
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
