package outlay

import (
	"time"

	"gioui.org/op"
)

// Animation holds state for an Animation between two states that
// is not invertible.
type Animation struct {
	time.Duration
	StartTime time.Time
}

// Progress returns the current progress through the animation
// as a value in the range [0,1]
func (n *Animation) Progress(g Gx) float32 {
	if n.Duration == time.Duration(0) {
		return 0
	}
	progressDur := g.Now.Sub(n.StartTime)
	if progressDur > n.Duration {
		return 1
	}
	g.Execute(op.InvalidateCmd{})
	progress := float32(progressDur.Milliseconds()) / float32(n.Duration.Milliseconds())
	return progress
}

func (n *Animation) Start(now time.Time) {
	n.StartTime = now
}

func (n *Animation) SetDuration(d time.Duration) {
	n.Duration = d
}

func (n *Animation) Animating(g Gx) bool {
	if n.Duration == 0 {
		return false
	}
	if g.Now.After(n.StartTime.Add(n.Duration)) {
		return false
	}
	return true
}
