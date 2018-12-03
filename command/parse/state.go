package parse

import (
	"strings"

	"github.com/raviqqe/jsonxx/command/debug"
	"github.com/raviqqe/parcom"
)

type state struct {
	*parcom.State
	filename, source string
}

func newState(f, s string) *state {
	return &state{parcom.NewState(s), f, s}
}

func (s state) debugInformation() *debug.Information {
	return debug.NewInformation(
		s.filename,
		s.LineNumber(),
		s.CharacterPosition(),
		strings.Split(s.source, "\n")[s.LineNumber()-1],
	)
}
