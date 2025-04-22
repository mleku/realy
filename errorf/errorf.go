// Package errorf is a convenience shortcut to use shorter names to access the lol.Logger.
package errorf

import (
	"realy.mleku.dev/lol"
)

var F, E, W, I, D, T lol.Err

func init() {
	F, E, W, I, D, T = lol.Main.Errorf.F, lol.Main.Errorf.E, lol.Main.Errorf.W, lol.Main.Errorf.I, lol.Main.Errorf.D, lol.Main.Errorf.T
}
