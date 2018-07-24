package action

import (
	"os"
	"path/filepath"
)

// Action knows everything to run gchl CLI actions
type Action struct {
	Name string
}

// New returns a new Action wrapper
func New() *Action {
	return newAction()
}

func newAction() *Action {
	name := "gchl"
	if len(os.Args) > 0 {
		name = filepath.Base(os.Args[0])
	}

	act := &Action{
		Name: name,
	}
	return act
}
