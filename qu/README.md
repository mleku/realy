# qu

##### observable signal channels

simple channels that act as breakers or momentary one-shot triggers.

can enable logging to get detailed information on channel state, and channels do
not panic if closed channels are attempted to be closed or signalled with.

provides a neat function based syntax for usage.

wait function does require use of the `<-` receive operator prefix to be used in
a select statement.

## usage

### creating channels:

#### unbuffered

    newSigChan := qu.T()

#### buffered

    newBufferedSigChan := qu.Ts(5)

#### closing

    newSigChan.Q()

#### signalling

    newBufferedSigChan.Signal()

#### logging features

    numberOpenUnbufferedChannels := GetOpenUnbufferedChanCount()
    
    numberOpenBufferedChannels := GetOpenBufferedChanCount()

print a list of closed and open channels known by qu:

    PrintChanState() 

## garbage collection

this library automatically cleans up closed channels once a minute to free
resources that have become unused.