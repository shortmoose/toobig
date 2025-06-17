package base

import "github.com/shortmoose/toobig/internal/config"

// Context is the basic structure that toobig uses to keep state.
// Many methods will take this structure as their first argument.
type Context struct {
	Command    string
	ConfigPath string

	config.TooBig

	Verbose       bool
	UpdateIsError bool
}
