package lol

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"

	"realy.lol/atomic"
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
	// LevelPrinter defines a set of terminal printing primitives that output with
	// extra data, time, log logLevelList, and code location

	// Ln prints lists of interfaces with spaces in between
	Ln func(a ...interface{})
	// F prints like fmt.Println surrounded by log details
	F func(format string, a ...interface{})
	// S prints a spew.Sdump for an enveloper slice
	S func(a ...interface{})
	// C accepts a function so that the extra computation can be avoided if it is
	// not being viewed
	C func(closure func() string)
	// Chk is a shortcut for printing if there is an error, or returning true
	Chk func(e er) bo
	// Err is a pass-through function that uses fmt.Errorf to construct an error
	// and returns the error after printing it to the log
	Err          func(format string, a ...interface{}) er
	LevelPrinter struct {
		Ln
		F
		S
		C
		Chk
		Err
	}
	LevelSpec struct {
		ID        no
		Name      string
		Colorizer func(a ...interface{}) string
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
	// sep is just a convenient shortcut for this very longwinded expression
	sep = string(os.PathSeparator)
	// writer can be swapped out for any io.*writer* that you want to use instead of
	// stdout.
	writer io.Writer = os.Stderr
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
	NoTimeStomp atomic.Bool
)

func NoSprint(a ...any) string { return "" }

// Log is a set of log printers for the various Level items.
type Log struct {
	F, E, W, I, D, T LevelPrinter
}

type Check struct {
	F, E, W, I, D, T Chk
}
type Errorf struct {
	F, E, W, I, D, T Err
}

type Logger struct {
	*Log
	*Check
	*Errorf
}

var Level atomic.Int32
var Main = &Logger{}

func init() {
	// Main = &Logger{}
	Main.Log, Main.Check, Main.Errorf = New(os.Stderr)
	SetLoggers(Info)
}

func SetLoggers(level no) {
	Main.Log.T.F("log level %s", LevelSpecs[level].Colorizer(LevelNames[level]))
	Level.Store(int32(level))
}

func GetLogLevel(level string) (i no) {
	for i = range LevelNames {
		if level == LevelNames[i] {
			return i
		}
	}
	return Info
}

func SetLogLevel(level string) {
	for i := range LevelNames {
		if level == LevelNames[i] {
			SetLoggers(i)
			return
		}
	}
}

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

func GetPrinter(l int32, writer io.Writer) LevelPrinter {
	return LevelPrinter{
		Ln: func(a ...interface{}) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(Timestamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				JoinStrings(a...),
				msgCol(GetLoc(2)),
			)
		},
		F: func(format string, a ...interface{}) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(Timestamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				fmt.Sprintf(format, a...),
				msgCol(GetLoc(2)),
			)
		},
		S: func(a ...interface{}) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(Timestamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				spew.Sdump(a...),
				msgCol(GetLoc(2)),
			)
		},
		C: func(closure func() string) {
			if Level.Load() < l {
				return
			}
			fmt.Fprintf(writer,
				"%s%s %s %s\n",
				msgCol(Timestamper()),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				closure(),
				msgCol(GetLoc(2)),
			)
		},
		Chk: func(e er) bo {
			if Level.Load() < l {
				return e != nil
			}
			if e != nil {
				fmt.Fprintf(writer,
					"%s%s %s %s\n",
					msgCol(Timestamper()),
					LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
					e.Error(),
					msgCol(GetLoc(2)),
				)
				return true
			}
			return false
		},
		Err: func(format string, a ...interface{}) er {
			if Level.Load() < l {
				fmt.Fprintf(writer,
					"%s%s %s %s\n",
					msgCol(Timestamper()),
					LevelSpecs[l].Colorizer(LevelSpecs[l].Name, " "),
					fmt.Sprintf(format, a...),
					msgCol(GetLoc(2)),
				)
			}
			return fmt.Errorf(format, a...)
		},
	}
}

func GetNullPrinter() LevelPrinter {
	return LevelPrinter{
		Ln:  func(a ...interface{}) {},
		F:   func(format string, a ...interface{}) {},
		S:   func(a ...interface{}) {},
		C:   func(closure func() string) {},
		Chk: func(e er) bo { return e != nil },
		Err: func(format string,
			a ...interface{}) er {
			return fmt.Errorf(format, a...)
		},
	}
}

func New(writer io.Writer) (l *Log, c *Check, errorf *Errorf) {
	l = &Log{
		T: GetPrinter(Trace, writer),
		D: GetPrinter(Debug, writer),
		I: GetPrinter(Info, writer),
		W: GetPrinter(Warn, writer),
		E: GetPrinter(Error, writer),
		F: GetPrinter(Fatal, writer),
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

// Timestamper e
func Timestamper() (s string) {
	if NoTimeStomp.Load() {
		return
	}
	timeText := fmt.Sprint(time.Now().UnixNano())
	lt := len(timeText)
	lb := lt + 1
	var timeBytes = make(by, lb)
	copy(timeBytes[lb-9:lb], timeText[lt-9:lt])
	timeBytes[lb-10] = '.'
	lb -= 10
	lt -= 9
	copy(timeBytes[:lb], timeText[:lt])
	return fmt.Sprint(string(timeBytes), " ")
}

var wd, _ = os.Getwd()

func GetNLoc(n no) (output string) {
	for ; n > 1; n-- {
		output += fmt.Sprintf("%s\n", GetLoc(n))
	}
	return
}

func GetLoc(skip no) (output string) {
	_, file, line, _ := runtime.Caller(skip)
	split := strings.Split(file, wd+string(os.PathSeparator))
	if len(split) < 2 {
		output = fmt.Sprintf("%s:%d", file, line)
	} else {
		output = fmt.Sprintf("%s:%d", split[1], line)
	}
	return
}
