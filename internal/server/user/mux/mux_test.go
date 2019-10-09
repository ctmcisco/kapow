// +build !race

package mux

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestSwappableMuxGetReturnsTheCurrentMux(t *testing.T) {
	sm := swappableMux{}
	mux := sm.get()
	if !reflect.DeepEqual(mux, sm.root) {
		t.Errorf("Returned mux is not the same %#v", mux)
	}
}

func TestSwappableMuxGetReturnsADifferentInstance(t *testing.T) {
	sm := swappableMux{}
	mux := sm.get()
	if &mux == &sm.root {
		t.Error("Returned mux is the same instance")
	}
}

func TestSwappableMuxGetWaitsForTheMutexToBeReleased(t *testing.T) {
	sm := swappableMux{}

	sm.m.Lock()
	defer sm.m.Unlock()

	c := make(chan *mux.Router)
	go func() { c <- sm.get() }()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-c:
		t.Error("Didn't acquire the mutex")
	default:
	}
}

func TestSwappableMuxGetIsAbleToReadWhileOthersAreReading(t *testing.T) {
	sm := swappableMux{}

	sm.m.RLock()
	defer sm.m.RUnlock()

	c := make(chan *mux.Router)
	go func() { c <- sm.get() }()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-c:
	default:
		t.Error("The mutex cannot be acquired")
	}
}

func TestSwappableMuxSetSetsTheGivenMux(t *testing.T) {
	sm := swappableMux{}
	m := mux.NewRouter()
	// nolint
	m.KeepContext = true

	sm.set(m)

	// nolint
	if !sm.root.KeepContext {
		t.Error("mux not set")
	}
}

func TestSwappableMuxSetSetsTheSameInstance(t *testing.T) {
	sm := swappableMux{}
	m := mux.NewRouter()

	sm.set(m)

	if m != sm.root {
		t.Error("Set mux is not the same instance")
	}
}

func TestSwappableMuxSetWaitsForWriterToReleaseMutex(t *testing.T) {
	sm := swappableMux{}

	sm.m.Lock()
	defer sm.m.Unlock()

	c := make(chan bool)
	go func() { sm.set(mux.NewRouter()); c <- true }()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-c:
		t.Error("Didn't acquire the mutex")
	default:
	}
}

func TestSwappableMuxSetWaitsForReadersToReleaseMutex(t *testing.T) {
	sm := swappableMux{}

	sm.m.RLock()
	defer sm.m.RUnlock()

	c := make(chan bool)
	go func() { sm.set(mux.NewRouter()); c <- true }()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-c:
		t.Error("Didn't acquire the mutex")
	default:
	}
}

func TestServeHTTPCallsInnerMux(t *testing.T) {
	called := false

	m := mux.NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { called = true })

	sm := swappableMux{root: m}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	sm.ServeHTTP(w, req)

	if !called {
		t.Error("Inner mux wasn't called")
	}
}

func TestServeHTTPCanServeWhenMuxIsReadLocked(t *testing.T) {
	called := false

	m := mux.NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { called = true })

	sm := swappableMux{root: m}
	sm.m.RLock()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	go sm.ServeHTTP(w, req)

	time.Sleep(10 * time.Millisecond)

	if !called {
		t.Error("Inner mux not called while mutex is read locked")
	}
}

func TestServeHTTPCallsInnerMuxAfterAcquiringLock(t *testing.T) {
	called := false

	m := mux.NewRouter()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { called = true })

	sm := swappableMux{root: m}
	sm.m.Lock()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	go sm.ServeHTTP(w, req)

	time.Sleep(10 * time.Millisecond)

	if called {
		t.Fatal("Mutex not acquired")
	}

	sm.m.Unlock()

	time.Sleep(10 * time.Millisecond)

	if !called {
		t.Error("Inner mux wasn't called after mutex released")
	}
}
