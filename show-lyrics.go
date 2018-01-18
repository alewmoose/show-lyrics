package main

import (
	"errors"
	"github.com/alewmoose/show-lyrics/cache"
	"github.com/alewmoose/show-lyrics/fetcher/azlyrics"
	"github.com/alewmoose/show-lyrics/fetcher/lyrics_wikia"
	"github.com/alewmoose/show-lyrics/player/cmus"
	"github.com/alewmoose/show-lyrics/player/mocp"
	"github.com/alewmoose/show-lyrics/songinfo"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
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
	lockFile := path.Join(dotDir, "lockfile")

	for _, dir := range []string{dotDir, cacheDir} {
		err := mkdirUnlessExists(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	flockErr := tryFlock(lockFile)
	if flockErr != nil {
		log.Fatalf("Failed to obtain a lock: %s", flockErr)
	}
	defer os.Remove(lockFile)

	client := &http.Client{}

	err := mainLoop(client, cacheDir)
	if err != nil {
		log.Fatal(err)
	}
}

func tryFlock(lockFile string) error {
	f, openErr := os.OpenFile(lockFile, os.O_CREATE, 0644)
	if openErr != nil {
		return openErr
	}

	flockErr := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if flockErr != nil {
		return flockErr
	}

	return nil
}

func mainLoop(client *http.Client, cacheDir string) error {
	signal.Ignore(syscall.SIGINT)

	songInfo, err := getSongInfo()
	if err != nil {
		return err
	}

	filePath, err := saveLyrics(client, cacheDir, songInfo)
	if err != nil {
		return err
	}

	cmdErr := make(chan error)
	defer close(cmdErr)
	cmd, err := startLess(filePath)
	if err != nil {
		return err
	}
	go waitCmd(cmd, cmdErr)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	prevSongInfo := *songInfo

LOOP:
	for {
		select {
		case <-ticker.C:
			songInfo, err = getSongInfo()
			if err != nil {
				break LOOP
			}
			if *songInfo == prevSongInfo {
				continue
			}
			prevSongInfo = *songInfo
			filePath, err = saveLyrics(client, cacheDir, songInfo)
			if err != nil {
				break LOOP
			}
			err = cmd.Process.Signal(syscall.SIGINT)
			if err != nil {
				break LOOP
			}
			err = <-cmdErr
			if err != nil {
				break LOOP
			}
			cmd, err = startLess(filePath)
			if err != nil {
				break LOOP
			}
			go waitCmd(cmd, cmdErr)
		case err = <-cmdErr:
			break LOOP
		}
	}

	return err
}

func saveLyrics(client *http.Client, cacheDir string, si *songinfo.SongInfo) (string, error) {
	lyricsCache := cache.New(cacheDir, si)
	if lyricsCache.Exists() {
		return lyricsCache.FilePath(), nil
	}

	lyrics, err := fetchLyrics(client, si)
	if err != nil {
		return "", err
	}
	lyrics = prepareLyrics(si, lyrics)
	err = lyricsCache.Store(lyrics)
	if err != nil {
		return "", err
	}

	return lyricsCache.FilePath(), nil
}

func startLess(filePath string) (*exec.Cmd, error) {
	cmd := exec.Command("less", "--clear-screen", "--quit-on-intr", "--ignore-case", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	return cmd, err
}

func waitCmd(cmd *exec.Cmd, res chan<- error) {
	state, err := cmd.Process.Wait()
	if state.Exited() {
		res <- nil
		return
	}
	if err != nil {
		res <- err
		return
	}
	res <- nil
}

var parensRe = regexp.MustCompile(`\(.+\)$`)

var lyricsFetchers = [...]func(*http.Client, *songinfo.SongInfo) ([]byte, error){
	azlyrics.Fetch,
	lyrics_wikia.Fetch,
}
var lfi = 0

func fetchLyrics(c *http.Client, si *songinfo.SongInfo) ([]byte, error) {
	for n := 0; n < len(lyricsFetchers); n++ {
		fetcher := lyricsFetchers[lfi]
		lfi = (lfi + 1) % len(lyricsFetchers)

		lyrics, err := fetcher(c, si)
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
	return nil, errors.New("Lyrics not found")
}

var SIGetters = [...]func() (*songinfo.SongInfo, error){
	cmus.GetSongInfo,
	mocp.GetSongInfo,
}

func getSongInfo() (*songinfo.SongInfo, error) {
	for i, getSI := range SIGetters {
		si, err := getSI()
		if err != nil {
			continue
		}
		if i != 0 {
			// next time try the active player first
			SIGetters[0], SIGetters[i] = SIGetters[i], SIGetters[0]
		}
		return si, nil
	}
	return nil, errors.New("No players running")
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
