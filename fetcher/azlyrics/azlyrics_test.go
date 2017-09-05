package azlyrics

import (
	"bytes"
	"errors"
	"github.com/alewmoose/show-lyrics/songinfo"
	"net/http"
	"strings"
	"testing"
)

type test struct {
	si   songinfo.SongInfo
	err  error
	text []byte
}

var tests = [...]test{
	{
		si: songinfo.SongInfo{
			Artist: "Mastodon",
			Title:  "Asleep In The Deep",
		},
		err: nil,
		text: []byte(strings.Join([]string{
			"The moment you walked in the room, my friend",
			"the demons, they all went away",
			"be careful, they're only asleep for a while",
			"pretending there's nothing to say",
			"",
			"Throw salt in all the corners here",
			"make sure you watch him leave",
			"",
			"Build up the walls around this house",
			"and dig out the rot in the floor",
			"block out the entrance with brick and stone",
			"and mortar that's made from coal",
			"",
			"Crawl into this hole I've made",
			"transform these feelings of fear",
			"",
			"I'm on fire",
			"say you'll remember her voice",
			"and I can't get you out of my mind",
			"",
			"Loose lips have fallen on deaf ears",
			"loose lips have fallen on blind eyes",
			"",
			"An ocean of sorrow surrounds this home",
			"I hope that we make it to shore",
			"as time chips away at the fortress walls",
			"it seems that we weathered the storm",
			"",
			"The sun begins to show itself",
			"revealing victory",
			"",
			"I'm on fire",
			"say you'll remember her voice",
			"and I can't get you out of my mind",
		}, "\n")),
	},
	{
		si: songinfo.SongInfo{
			Artist: "naosehntaoshftrdru",
			Title:  "aosehntaoshftrdrutn",
		},
		err:  errors.New("404 Not Found"),
		text: []byte{},
	},
}

func TestFetch(t *testing.T) {
	client := &http.Client{}
	fmt := "Failed test #%d:\ngot:\n\"%s\"\nexpected:\n\"%s\"\n"

	for i, fetchTest := range tests {
		text, err := Fetch(client, &fetchTest.si)

		if !bytes.Equal(text, fetchTest.text) {
			t.Errorf(fmt, i+1, string(text), string(fetchTest.text))
		}

		var gotErr, expErr string
		if err != nil {
			gotErr = err.Error()
		}
		if fetchTest.err != nil {
			expErr = fetchTest.err.Error()
		}
		if gotErr != expErr {
			t.Errorf(fmt, i+1, gotErr, expErr)
		}
	}
}
