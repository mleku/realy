package chaincfg

import (
	"fmt"
	"time"

	"realy.lol/ec/wire"
)

var (
	// ErrNoBlockClock is returned when an operation fails due to lack of
	// synchornization with the current up to date block clock.
	ErrNoBlockClock = fmt.Errorf("no block clock synchronized")
)

// ConsensusDeploymentStarter determines if a given consensus deployment has
// started. A deployment has started once according to the current "time", the
// deployment is eligible for activation once a perquisite condition has
// passed.
type ConsensusDeploymentStarter interface {
	// HasStarted returns true if the consensus deployment has started.
	HasStarted(*wire.BlockHeader) (bo, er)
}

// ConsensusDeploymentEnder determines if a given consensus deployment has
// ended. A deployment has ended once according got eh current "time", the
// deployment is no longer eligible for activation.
type ConsensusDeploymentEnder interface {
	// HasEnded returns true if the consensus deployment has ended.
	HasEnded(*wire.BlockHeader) (bo, er)
}

// BlockClock is an abstraction over the past median time computation. The past
// median time computation is used in several consensus checks such as CSV, and
// also BIP 9 version bits. This interface allows callers to abstract away the
// computation of the past median time from the perspective of a given block
// header.
type BlockClock interface {
	// PastMedianTime returns the past median time from the PoV of the
	// passed block header. The past median time is the median time of the
	// 11 blocks prior to the passed block header.
	PastMedianTime(*wire.BlockHeader) (time.Time, er)
}

// ClockConsensusDeploymentEnder is a more specialized version of the
// ConsensusDeploymentEnder that uses a BlockClock in order to determine if a
// deployment has started or not.
//
// NOTE: Any calls to HasEnded will _fail_ with ErrNoBlockClock if they
// happen before SynchronizeClock is executed.
type ClockConsensusDeploymentEnder interface {
	ConsensusDeploymentEnder
	// SynchronizeClock synchronizes the target ConsensusDeploymentStarter
	// with the current up-to date BlockClock.
	SynchronizeClock(clock BlockClock)
}

// MedianTimeDeploymentStarter is a ClockConsensusDeploymentStarter that uses
// the median time past of a target block node to determine if a deployment has
// started.
type MedianTimeDeploymentStarter struct {
	blockClock BlockClock
	startTime  time.Time
}

// NewMedianTimeDeploymentStarter returns a new instance of a
// MedianTimeDeploymentStarter for a given start time. Using a time.Time
// instance where IsZero() is true, indicates that a deployment should be
// considered to always have been started.
func NewMedianTimeDeploymentStarter(startTime time.Time) *MedianTimeDeploymentStarter {
	return &MedianTimeDeploymentStarter{
		startTime: startTime,
	}
}

// HasStarted returns true if the consensus deployment has started.
func (m *MedianTimeDeploymentStarter) HasStarted(blkHeader *wire.BlockHeader) (bo,
	er) {
	switch {
	// If we haven't yet been synchronized with a block clock, then we
	// can't tell the time, so we'll fail.
	case m.blockClock == nil:
		return false, ErrNoBlockClock
	// If the time is "zero", then the deployment has always started.
	case m.startTime.IsZero():
		return true, nil
	}
	medianTime, err := m.blockClock.PastMedianTime(blkHeader)
	if err != nil {
		return false, err
	}
	// We check both after and equal here as after will fail for equivalent
	// times, and we want to be inclusive.
	return medianTime.After(m.startTime) || medianTime.Equal(m.startTime), nil
}

// MedianTimeDeploymentEnder is a ClockConsensusDeploymentEnder that uses the
// median time past of a target block to determine if a deployment has ended.
type MedianTimeDeploymentEnder struct {
	blockClock BlockClock
	endTime    time.Time
}

// NewMedianTimeDeploymentEnder returns a new instance of the
// MedianTimeDeploymentEnder anchored around the passed endTime.  Using a
// time.Time instance where IsZero() is true, indicates that a deployment
// should be considered to never end.
func NewMedianTimeDeploymentEnder(endTime time.Time) *MedianTimeDeploymentEnder {
	return &MedianTimeDeploymentEnder{
		endTime: endTime,
	}
}

// HasEnded returns true if the deployment has ended.
func (m *MedianTimeDeploymentEnder) HasEnded(blkHeader *wire.BlockHeader) (bo,
	er) {
	switch {
	// If we haven't yet been synchronized with a block clock, then we can't tell
	// the time, so we'll we haven't yet been synchronized with a block
	// clock, then w can't tell the time, so we'll fail.
	case m.blockClock == nil:
		return false, ErrNoBlockClock
	// If the time is "zero", then the deployment never ends.
	case m.endTime.IsZero():
		return false, nil
	}
	medianTime, err := m.blockClock.PastMedianTime(blkHeader)
	if err != nil {
		return false, err
	}
	// We check both after and equal here as after will fail for equivalent
	// times, and we want to be inclusive.
	return medianTime.After(m.endTime) || medianTime.Equal(m.endTime), nil
}

// EndTime returns the raw end time of the deployment.
func (m *MedianTimeDeploymentEnder) EndTime() time.Time {
	return m.endTime
}

// SynchronizeClock synchronizes the target ConsensusDeploymentEnder with the
// current up-to date BlockClock.
func (m *MedianTimeDeploymentEnder) SynchronizeClock(clock BlockClock) {
	m.blockClock = clock
}

// A compile-time assertion to ensure MedianTimeDeploymentEnder implements the
// ClockConsensusDeploymentStarter interface.
var _ ClockConsensusDeploymentEnder = (*MedianTimeDeploymentEnder)(nil)
