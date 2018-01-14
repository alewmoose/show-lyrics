package lyrics_wikia

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
			Artist: "Mogwai",
			Title:  "Take Me Somewhere Nice",
		},
		err: nil,
		text: []byte(strings.Join([]string{
			"Ghosts in the photograph",
			"Never lied to me",
			"I'd be all of that",
			"I'd be all of that",
			"",
			"A false memory",
			"Would be everything",
			"A denial",
			"My eliminent",
			"",
			"What was that for?",
			"What was that for?",
			"",
			"What would you do",
			"If you saw spaceships",
			"Over Glasgow?",
			"Would you fear them?",
			"",
			"Every aircraft",
			"Every camera",
			"Is a wish that",
			"Wasn't granted",
			"",
			"What was that for?",
			"What was that for?",
			"",
			"Try to be bad",
			"Try to be bad",
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
