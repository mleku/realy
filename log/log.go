// Package log is a convenience shortcut to use shorter names to access the lol.Logger.
package log

import (
	"realy.mleku.dev/lol"
)

var F, E, W, I, D, T lol.LevelPrinter

func init() {
	F, E, W, I, D, T = lol.Main.Log.F, lol.Main.Log.E, lol.Main.Log.W, lol.Main.Log.I, lol.Main.Log.D, lol.Main.Log.T
}
