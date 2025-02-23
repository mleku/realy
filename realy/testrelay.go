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
	started := make(chan bool)
	go srv.Start("127.0.0.1", 0, started)
	<-started
	return srv
}

type testRelay struct {
	c           context.T
	Cancel      context.F
	name        string
	storage     store.I
	init        func() error
	onShutdown  func(context.T)
	acceptEvent func(*event.T) bool
}

func (tr *testRelay) Name() string     { return tr.name }
func (tr *testRelay) Storage() store.I { return tr.storage }
func (tr *testRelay) Origin() string   { return "example.com" }
func (tr *testRelay) Init() error {
	tr.c, tr.Cancel = context.Cancel(context.Bg())
	if fn := tr.init; fn != nil {
		return fn()
	}
	return nil
}

func (tr *testRelay) NoLimiter(pubKey []byte) bool {
	return false
}

func (tr *testRelay) Owners() [][]byte { return nil }

func (tr *testRelay) OnShutdown(c context.T) {
	if fn := tr.onShutdown; fn != nil {
		fn(c)
	}
}

func (tr *testRelay) AcceptEvent(c context.T, evt *event.T, hr *http.Request, origin string,
	authedPubkey []byte) (ok bool, notice string, after func()) {
	if fn := tr.acceptEvent; fn != nil {
		return fn(evt), "", nil
	}
	return true, "", nil
}

type testStorage struct {
	init        func() error
	close       func()
	queryEvents func(context.T, *filter.T) ([]*event.T, error)
	deleteEvent func(c context.T, eid *eventid.T, noTombstone ...bool) error
	saveEvent   func(context.T, *event.T) error
	countEvents func(context.T, *filter.T) (int, bool, error)
}

func (string *testStorage) Import(r io.Reader) {
	// TODO implement me
	panic("implement me")
}

func (string *testStorage) Export(c context.T, w io.Writer, pubkeys ...[]byte) {
	// TODO implement me
	panic("implement me")
}

func (string *testStorage) Sync() (err error) {
	// TODO implement me
	panic("implement me")
}

func (string *testStorage) Nuke() (err error) {
	// TODO implement me
	panic("implement me")
}

func (string *testStorage) Path() string {
	// TODO implement me
	panic("implement me")
}

func (string *testStorage) Init(path string) error {
	if fn := string.init; fn != nil {
		return fn()
	}
	return nil
}

func (string *testStorage) Close() (err error) {
	if fn := string.close; fn != nil {
		fn()
	}
	return
}

func (string *testStorage) QueryEvents(c context.T, f *filter.T) (evs event.Ts, err error) {
	if fn := string.queryEvents; fn != nil {
		return fn(c, f)
	}
	return nil, nil
}

func (string *testStorage) DeleteEvent(c context.T, evt *eventid.T, noTombstone ...bool) error {
	if fn := string.deleteEvent; fn != nil {
		return fn(c, evt)
	}
	return nil
}

func (string *testStorage) SaveEvent(c context.T, e *event.T) error {
	if fn := string.saveEvent; fn != nil {
		return fn(c, e)
	}
	return nil
}

func (string *testStorage) CountEvents(c context.T, f *filter.T) (int, bool, error) {
	if fn := string.countEvents; fn != nil {
		return fn(c, f)
	}
	return 0, false, nil
}
