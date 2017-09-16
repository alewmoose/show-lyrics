package cmus

import (
	"errors"
	"github.com/alewmoose/show-lyrics/songinfo"
	"os/exec"
	"regexp"
)

func GetSongInfo() (*songinfo.SongInfo, error) {
	stats, err := getStats()
	if err != nil {
		return nil, err
	}
	Songinfo, err := parseStats(stats)
	if err != nil {
		return nil, err
	}
	return Songinfo, nil
}

// TODO: error
func getStats() ([]byte, error) {
	cmd := exec.Command("cmus-remote", "-Q")
	return cmd.Output()
}

var artistRe = regexp.MustCompile(`(?m)^tag\s+artist\s+(.+)\s*$`)
var albumArtistRe = regexp.MustCompile(`(?m)^tag\s+albumartist\s+(.+)\s*$`)
var titleRe = regexp.MustCompile(`(?m)^tag\s+title\s+(.+)\s*$`)

func regexpMatch(re *regexp.Regexp, buf []byte) []byte {
	match := re.FindAllSubmatch(buf, 1)
	if len(match) > 0 {
		return match[0][1]
	}
	return nil
}

func parseStats(stats []byte) (*songinfo.SongInfo, error) {
	artist := regexpMatch(artistRe, stats)
	if artist == nil {
		artist = regexpMatch(albumArtistRe, stats)
	}
	title := regexpMatch(titleRe, stats)

	if artist == nil || title == nil {
		return nil, errors.New("Failed to parse cmus status")
	}

	si := songinfo.SongInfo{
		Artist: string(artist),
		Title:  string(title),
	}

	return &si, nil
}
