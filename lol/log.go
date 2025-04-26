// Package lol (log of location) is a simple logging library that prints a high precision unix
// timestamp and the source location of a log print to make tracing errors simpler. Includes a
// set of logging levels and the ability to filter out higher log levels for a more quiet
// output.
package lol

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
)

const (
	Off = iota
	Fatal
	Error
	Warn
	Info
	Debug
	Trace
)

var LevelNames = []string{
	"off",
	"fatal",
	"error",
	"warn",
	"info",
	"debug",
	"trace",
}

type (
	// LevelPrinter defines a set of terminal printing primitives that output with extra data,
	// time, log logLevelList, and code location

	// Ln prints lists of interfaces with spaces in between
	Ln func(a ...interface{})
	// F prints like fmt.Println surrounded []byte log details
	F func(format string, a ...interface{})
	// S prints a spew.Sdump for an enveloper slice
	S func(a ...interface{})
	// C accepts a function so that the extra computation can be avoided if it is not being
	// viewed
	C func(closure func() string)
	// Chk is a shortcut for printing if there is an error, or returning true
	Chk func(e error) bool
	// Err is a pass-through function that uses fmt.Errorf to construct an error and returns the
	// error after printing it to the log
	Err func(format string, a ...any) error

	// LevelPrinter is the set of log printers on each log level.
	LevelPrinter struct {
		Ln
		F
		S
		C
		Chk
		Err
	}

	// LevelSpec is the name, ID and Colorizer for a log level.
	LevelSpec struct {
		ID        int
		Name      string
		Colorizer func(a ...any) string
	}

	// Entry is a log entry to be printed as json to the log file
	Entry struct {
		Time         time.Time
		Level        string
		Package      string
		CodeLocation string
		Text         string
	}
)

var (
	// Writer can be swapped out for any io.*Writer* that you want to use instead of stdout.
	Writer io.Writer = os.Stderr

	// LevelSpecs specifies the id, string name and color-printing function
	LevelSpecs = []LevelSpec{
		{Off, "", NoSprint},
		{Fatal, "FTL", color.New(color.BgRed, color.FgHiWhite).Sprint},
		{Error, "ERR", color.New(color.FgHiRed).Sprint},
		{Warn, "WRN", color.New(color.FgHiYellow).Sprint},
		{Info, "INF", color.New(color.FgHiGreen).Sprint},
		{Debug, "DBG", color.New(color.FgHiBlue).Sprint},
		{Trace, "TRC", color.New(color.FgHiMagenta).Sprint},
	}
	NoTimeStamp atomic.Bool
	ShortLoc    atomic.Bool
)

// NoSprint is a noop for sprint (it returns nothing no matter what is given to it).
func NoSprint(a ...any) string { return "" }

// Log is a set of log printers for the various Level items.
type Log struct {
	F, E, W, I, D, T LevelPrinter
}

// Check is the set of log levels for a Check operation (prints an error if the error is not
// nil).
type Check struct {
	F, E, W, I, D, T Chk
}

// Errorf prints an error that is also returned as an error, so the error is logged at the site.
type Errorf struct {
	F, E, W, I, D, T Err
}

// Logger is a collection of things that creates a logger, including levels.
type Logger struct {
	*Log
	*Check
	*Errorf
}

// Level is the level that the logger is printing at.
var Level atomic.Int32

// Main is the main logger.
var Main = &Logger{}

func init() {
	// Main = &Logger{}
	Main.Log, Main.Check, Main.Errorf = New(os.Stderr, 2)
	SetLoggers(Info)
}

// SetLoggers configures a log level.
func SetLoggers(level int) {
	Main.Log.T.F("log level %s", LevelSpecs[level].Colorizer(LevelNames[level]))
	Level.Store(int32(level))
}

// GetLogLevel returns the log level number of a string log level.
func GetLogLevel(level string) (i int) {
	for i = range LevelNames {
		if level == LevelNames[i] {
			return i
		}
	}
	return Info
}

// SetLogLevel sets the log level of the logger.
func SetLogLevel(level string) {
	for i := range LevelNames {
		if level == LevelNames[i] {
			SetLoggers(i)
			return
		}
	}
	SetLoggers(Trace)
}

// JoinStrings joins together anything into a set of strings with space separating the items.
func JoinStrings(a ...any) (s string) {
	for i := range a {
		s += fmt.Sprint(a[i])
		if i < len(a)-1 {
			s += " "
		}
	}
	return
}

var msgCol = color.New(color.FgBlue).Sprint

// GetPrinter returns a full logger that writes to the provided io.Writer.
func GetPrinter(l int32, writer io.Writer, skip int) LevelPrinter {
	return LevelPrinter{
		Ln: func(a ...interface{}) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(TimeStamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				JoinStrings(a...),
				msgCol(GetLoc(skip)),
			)
		},
		F: func(format string, a ...interface{}) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(TimeStamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				fmt.Sprintf(format, a...),
				msgCol(GetLoc(skip)),
			)
		},
		S: func(a ...interface{}) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(TimeStamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				spew.Sdump(a...),
				msgCol(GetLoc(skip)),
			)
		},
		C: func(closure func() string) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(TimeStamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				closure(),
				msgCol(GetLoc(skip)),
			)
		},
		Chk: func(e error) bool {
			if Level.Load() < l {
				return e != nil
			}
			if e != nil {
				fmt.Fprintf(writer,
					"%s%s %s %s\n",
					msgCol(TimeStamper()),
					LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
					e.Error(),
					msgCol(GetLoc(skip)),
				)
				return true
			}
			return false
		},
		Err: func(format string, a ...interface{}) error {
			if Level.Load() < l {
				fmt.Fprintf(writer,
					"%s%s %s %s\n",
					msgCol(TimeStamper()),
					LevelSpecs[l].Colorizer(LevelSpecs[l].Name, " "),
					fmt.Sprintf(format, a...),
					msgCol(GetLoc(skip)),
				)
			}
			return fmt.Errorf(format, a...)
		},
	}
}

// GetNullPrinter is a logger that doesn't log.
func GetNullPrinter() LevelPrinter {
	return LevelPrinter{
		Ln:  func(a ...interface{}) {},
		F:   func(format string, a ...interface{}) {},
		S:   func(a ...interface{}) {},
		C:   func(closure func() string) {},
		Chk: func(e error) bool { return e != nil },
		Err: func(format string, a ...interface{}) error { return fmt.Errorf(format, a...) },
	}
}

// New creates a new logger with all the levels and things.
func New(writer io.Writer, skip int) (l *Log, c *Check, errorf *Errorf) {
	if writer == nil {
		writer = Writer
	}
	l = &Log{
		T: GetPrinter(Trace, writer, skip),
		D: GetPrinter(Debug, writer, skip),
		I: GetPrinter(Info, writer, skip),
		W: GetPrinter(Warn, writer, skip),
		E: GetPrinter(Error, writer, skip),
		F: GetPrinter(Fatal, writer, skip),
	}
	c = &Check{
		F: l.F.Chk,
		E: l.E.Chk,
		W: l.W.Chk,
		I: l.I.Chk,
		D: l.D.Chk,
		T: l.T.Chk,
	}
	errorf = &Errorf{
		F: l.F.Err,
		E: l.E.Err,
		W: l.W.Err,
		I: l.I.Err,
		D: l.D.Err,
		T: l.T.Err,
	}
	return
}

// TimeStamper generates the timestamp for logs.
func TimeStamper() (s string) {
	if NoTimeStamp.Load() {
		return
	}
	return time.Now().Format("2006-01-02T15:04:05Z07:00.000 ")
}

// var wd, _ = os.Getwd()

// GetNLoc returns multiple levels of depth of code location from the current.
func GetNLoc(n int) (output string) {
	for ; n > 1; n-- {
		output += fmt.Sprintf("%s\n", GetLoc(n))
	}
	return
}

var prefix string

func init() {
	// this enables us to remove the base of the path for a more compact code location string,
	// this can be used with tilix custom hyperlinks feature
	//
	// create a script called `setcurrent` in your PATH ( eg ~/.local/bin/setcurrent )
	//
	//   #!/usr/bin/bash
	//   echo $(pwd) > ~/.current
	//
	// set the following environment variable in your ~/.bashrc
	//
	//   export PROMPT_COMMAND='setcurrent'
	//
	// using the following regular expressions, replacing the path as necessary, and setting
	// perhaps a different program than ide (this is for goland, i use an alias to the binary)
	//
	//   ^((([a-zA-Z@0-9-_.]+/)+([a-zA-Z@0-9-_.]+)):([0-9]+))    ide --line $5 $(cat /home/mleku/.current)/$2
	//   [ ]((([a-zA-Z@0-9-_./]+)+([a-zA-Z@0-9-_.]+)):([0-9]+))  ide --line $5 $(cat /home/mleku/.current)/$2
	//   ([/](([a-zA-Z@0-9-_.]+/)+([a-zA-Z@0-9-_.]+)):([0-9]+))  ide --line $5 /$2
	//
	// and so long as you use this with an app containing /lol/log.go as this one is, this finds
	// that path and trims it off from the log line locations and in tilix you can click on the
	// file locations that are relative to the CWD where you are running the relay from. if this
	// is a remote machine, just go to the location where your source code is to make it work.
	//
	_, file, _, _ := runtime.Caller(0)
	prefix = file[:len(file)-10]
}

// GetLoc returns the code location of the caller.
func GetLoc(skip int) (output string) {
	_, file, line, _ := runtime.Caller(skip)
	if strings.Contains(file, "pkg/mod/") || !ShortLoc.Load() {
	} else {
		var split []string
		split = strings.Split(file, prefix)
		file = split[1]
	}
	output = fmt.Sprintf("%s:%d", file, line)
	return
}
