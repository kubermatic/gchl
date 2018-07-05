package git

import (
	"testing"
)

func TestHasReleaseNotes(t *testing.T) {
	want := true
	text := `___release-notes
This is a text
___`

	if got := hasReleaseNotes(text); got != want {
		t.Errorf("Error: got %v, want %v", got, want)
	}

}
