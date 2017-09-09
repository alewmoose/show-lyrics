package cache

import (
	"github.com/alewmoose/show-lyrics/songinfo"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type lyricsCache struct {
	artistDir, lyricsFile string
}

func New(cacheDir string, si *songinfo.SongInfo) *lyricsCache {
	artist := replaceSlashes(si.Artist)
	title := replaceSlashes(si.Title)
	artistDir := path.Join(cacheDir, artist)
	lyricsFile := path.Join(artistDir, title+".txt")
	return &lyricsCache{
		artistDir:  artistDir,
		lyricsFile: lyricsFile,
	}
}

func (c *lyricsCache) Exists() bool {
	_, err := os.Stat(c.lyricsFile)
	return err == nil
}

func (c *lyricsCache) FilePath() string {
	return c.lyricsFile
}

func (c *lyricsCache) Store(lyrics []byte) error {
	_, err := os.Stat(c.artistDir)
	if err != nil {
		err = os.Mkdir(c.artistDir, 0755)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(c.lyricsFile, lyrics, 0644)
}

func replaceSlashes(s string) string {
	return strings.Replace(s, "/", "_", -1)
}
