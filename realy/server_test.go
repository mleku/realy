package realy

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gobwas/ws/wsutil"

	"realy.lol/context"
	"realy.lol/ratel"
	"realy.lol/ws"
)

func TestServerStartShutdown(t *testing.T) {
	var (
		inited      bool
		storeInited bool
		shutdown    bool
	)
	c, cancel := context.Cancel(context.Bg())
	rl := &testRelay{
		Ctx:    c,
		Cancel: cancel,
		name:   "test server start",
		init: func() E {
			inited = true
			return nil
		},
		onShutdown: func(context.T) { shutdown = true },
		storage: &testStorage{
			init: func() E { storeInited = true; return nil },
		},
	}
	srv, _ := NewServer(ServerParams{
		Ctx:      c,
		Cancel:   cancel,
		Rl:       rl,
		MaxLimit: ratel.DefaultMaxLimit,
	})
	ready := make(chan bool)
	done := make(chan E)
	go func() {
		done <- srv.Start("127.0.0.1", 0, ready)
		close(done)
	}()
	<-ready

	// verify everything's initialized
	if !inited {
		t.Error("didn't call testRelay.init")
	}
	if !storeInited {
		t.Error("didn't call testStorage.init")
	}

	// check that http requests are served
	if _, err := http.Get("http://" + srv.Addr); chk.T(err) {
		t.Errorf("GET %s: %v", srv.Addr, err)
	}

	// verify server shuts down
	defer srv.Cancel()
	srv.Shutdown()
	if !shutdown {
		t.Error("didn't call testRelay.onShutdown")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("srv.Start: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("srv.Start too long to return")
	}
}

func TestServerShutdownWebsocket(t *testing.T) {
	// set up a new relay server
	srv := startTestRelay(context.Bg(), t, &testRelay{storage: &testStorage{}})

	// connect a client to it
	ctx1, cancel := context.Timeout(context.Bg(), 2*time.Second)
	defer cancel()
	client, err := ws.RelayConnect(ctx1, "ws://"+srv.Addr)
	if err != nil {
		t.Fatalf("nostr.RelayConnectContext: %v", err)
	}

	// now, shut down the server
	defer srv.Cancel()
	srv.Shutdown()

	// wait for the client to receive a "connection close"
	time.Sleep(1 * time.Second)
	err = client.ConnectionError
	if e := errors.Unwrap(err); e != nil {
		err = e
	}
	if _, ok := err.(wsutil.ClosedError); !ok {
		t.Errorf("client.ConnectionError: %v (%T); want wsutil.ClosedError", err, err)
	}
}
