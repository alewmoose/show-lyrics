package songinfo

type SongInfo struct {
	Artist, Title string
}

func (si *SongInfo) PrettyTitle() string {
	return si.Artist + " - " + si.Title
}
