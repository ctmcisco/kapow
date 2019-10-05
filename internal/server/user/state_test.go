// +build !race

package user

import (
	"github.com/BBVA/kapow/internal/server/model"
	"reflect"
	"testing"
	"time"
)

func TestNewReturnAnEmptyStruct(t *testing.T) {
	srl := New()

	if len(srl.rs) != 0 {
		t.Error("Unexpected member in slice")
	}
}

func TestPackageHaveASingletonEmptyRouteList(t *testing.T) {
	if !reflect.DeepEqual(Routes, New()) {
		t.Error("Routes is not an empty safeRouteList")
	}
}

func TestAppendAppendsANewRouteToTheList(t *testing.T) {
	srl := New()

	srl.Append(model.Route{})

	if len(srl.rs) == 0 {
		t.Error("Route not added to the list")
	}
}

func TestAppendAdquiresMutexBeforeAdding(t *testing.T) {
	srl := New()

	srl.m.Lock()
	defer srl.m.Unlock()
	go srl.Append(model.Route{})

	time.Sleep(10 * time.Millisecond)

	if len(srl.rs) != 0 {
		t.Error("Route added while mutex was adquired")
	}
}

func TestAppendAddsRouteAfterMutexIsReleased(t *testing.T) {
	srl := New()

	srl.m.Lock()
	go srl.Append(model.Route{})
	srl.m.Unlock()

	time.Sleep(10 * time.Millisecond)

	if len(srl.rs) != 1 {
		t.Error("Route not added after mutex release")
	}
}

func TestSnapshotReturnTheCurrentListOfRoutes(t *testing.T) {
	srl := New()
	srl.Append(model.Route{Id: "FOO"})

	rs := srl.Snapshot()

	if !reflect.DeepEqual(srl.rs, rs) {
		t.Error("Route list returned is not the current one")
	}
}

func TestSnapshotReturnADeepCopyOfTheListWhenEmpty(t *testing.T) {
	srl := New()

	rs := srl.Snapshot()

	if rs != nil {
		t.Fatal("Route list copy is not empty")
	}
}

func TestSnapshotReturnADeepCopyOfTheListWhenNonEmpty(t *testing.T) {
	srl := New()
	srl.Append(model.Route{Id: "FOO"})

	rs := srl.Snapshot()

	if &rs == &srl.rs {
		t.Fatal("Route list is not a copy")
	}

	for i := 0; i < len(rs); i++ {
		if &rs[i] == &srl.rs[i] {
			t.Errorf("Route %q is not a copy", i)
		}
	}
}

func TestSnapshotWaitsForTheWriterToFinish(t *testing.T) {
	srl := New()
	srl.Append(model.Route{Id: "FOO"})

	srl.m.Lock()
	defer srl.m.Unlock()

	c := make(chan []model.Route)
	go func() { c <- srl.Snapshot() }()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-c:
		t.Error("Route list readed while mutex was adquired")
	default: // This default prevents the select from being blocking
	}
}

func TestSnapshotNonBlockingReadWithOtherReaders(t *testing.T) {
	srl := New()
	srl.Append(model.Route{Id: "FOO"})

	srl.m.RLock()
	defer srl.m.RUnlock()

	c := make(chan []model.Route)
	go func() { c <- srl.Snapshot() }()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-c:
	default: // This default prevents the select from being blocking
		t.Error("Route list couldn't be readed while mutex was adquired for read")
	}
}
