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
			Artist: "A Perfect circle",
			Title:  "Breña",
		},
		err: nil,
		text: []byte(strings.Join([]string{
			"My reflection",
			"Wraps and pulls me under",
			"healing waters to be",
			"Bathed in Brena",
			"",
			"Guides me",
			"Safely in",
			"Worlds I've never been to",
			"Heal me",
			"Heal me",
			"My dear Brena",
			"",
			"So vulnerable",
			"But it's alright",
			"",
			"Heal me",
			"Heal me",
			"My dear Brena",
			"",
			"Show me lonely and",
			"Show me openings",
			"To lead me closer to you",
			"My dear Brena",
			"",
			"(Feeling so) vulnerable",
			"But it's alright",
			"",
			"Opening to... heal...",
			"Opening to... heal...",
			"Heal.. Heal.. Heal...",
			"",
			"Heal me",
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
	const fmt = "Failed test #%d:\ngot:\n\"%s\"\nexpected:\n\"%s\"\n"
	client := &http.Client{}

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
