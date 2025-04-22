// Package chk is a convenience shortcut to use shorter names to access the lol.Logger.
package chk

import (
	"realy.mleku.dev/lol"
)

var F, E, W, I, D, T lol.Chk

func init() {
	F, E, W, I, D, T = lol.Main.Check.F, lol.Main.Check.E, lol.Main.Check.W, lol.Main.Check.I,
		lol.Main.Check.D, lol.Main.Check.T
}
