//go:build linux

package interrupt

import (
	"os"
	"syscall"

	"github.com/kardianos/osext"

	"realy.mleku.dev/log"
)

// Restart uses syscall.Exec to restart the process. MacOS and Windows are not implemented,
// currently.
func Restart() {
	log.D.Ln("restarting")
	file, e := osext.Executable()
	if e != nil {
		log.E.Ln(e)
		return
	}
	e = syscall.Exec(file, os.Args, os.Environ())
	if e != nil {
		log.F.Ln(e)
	}
}
