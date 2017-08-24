package cmus

import (
	"errors"
	"github.com/alewmoose/show-lyrics/songinfo"
	"os/exec"
	"regexp"
)

func GetSongInfo() (*songinfo.SongInfo, error) {
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
