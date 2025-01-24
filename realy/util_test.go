package realy

import (
	"io"
	"net/http"
	"testing"

	"realy.lol/context"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/store"
	"realy.lol/units"
)

func startTestRelay(c context.T, t *testing.T, tr *testRelay) *Server {
	t.Helper()
	srv, _ := NewServer(&ServerParams{
		Ctx:      c,
		Cancel:   func() {},
		Rl:       tr,
		MaxLimit: 500 * units.Kb,
	})
	started := make(chan bo)
	go srv.Start("127.0.0.1", 0, started)
	<-started
	return srv
}

type testRelay struct {
	cx
	Cancel      context.F
	name        st
	storage     store.I
	init        func() er
	onShutdown  func(context.T)
	acceptEvent func(*event.T) bo
}

func (tr *testRelay) Name() st         { return tr.name }
func (tr *testRelay) Storage() store.I { return tr.storage }
func (tr *testRelay) Origin() st       { return "example.com" }
func (tr *testRelay) Init() er {
	tr.cx, tr.Cancel = context.Cancel(context.Bg())
	if fn := tr.init; fn != nil {
		return fn()
	}
	return nil
}

func (tr *testRelay) NoLimiter(pubKey by) bo {
	return false
}

func (tr *testRelay) OnShutdown(c context.T) {
	if fn := tr.onShutdown; fn != nil {
		fn(c)
	}
}

func (tr *testRelay) AcceptEvent(c context.T, evt *event.T, hr *http.Request, origin st,
	authedPubkey by) (ok bo, notice st, after func()) {
	if fn := tr.acceptEvent; fn != nil {
		return fn(evt), "", nil
	}
	return true, "", nil
}

type testStorage struct {
	init        func() er
	close       func()
	queryEvents func(context.T, *filter.T) ([]*event.T, er)
	deleteEvent func(c context.T, eid *eventid.T, noTombstone ...bo) er
	saveEvent   func(context.T, *event.T) er
	countEvents func(context.T, *filter.T) (no, bo, er)
}

func (st *testStorage) Import(r io.Reader) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Export(c cx, w io.Writer, pubkeys ...by) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Sync() (err er) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Nuke() (err er) {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Path() st {
	// TODO implement me
	panic("implement me")
}

func (st *testStorage) Init(path st) er {
	if fn := st.init; fn != nil {
		return fn()
	}
	return nil
}

func (st *testStorage) Close() (err er) {
	if fn := st.close; fn != nil {
		fn()
	}
	return
}

func (st *testStorage) QueryEvents(c context.T, f *filter.T) (evs event.Ts, err er) {
	if fn := st.queryEvents; fn != nil {
		return fn(c, f)
	}
	return nil, nil
}

func (st *testStorage) DeleteEvent(c context.T, evt *eventid.T, noTombstone ...bo) er {
	if fn := st.deleteEvent; fn != nil {
		return fn(c, evt)
	}
	return nil
}

func (st *testStorage) SaveEvent(c context.T, e *event.T) er {
	if fn := st.saveEvent; fn != nil {
		return fn(c, e)
	}
	return nil
}

func (st *testStorage) CountEvents(c context.T, f *filter.T) (no, bo, er) {
	if fn := st.countEvents; fn != nil {
		return fn(c, f)
	}
	return 0, false, nil
}
