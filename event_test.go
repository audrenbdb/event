package event_test

import (
	"context"
	"event"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestEmitter(t *testing.T) {
	em := event.NewEmitter[int](context.Background())

	l1 := em.NewListener()
	l2 := em.NewListener()
	l3 := em.NewListener(event.On(func(n int) bool {
		return n >= 100
	}))
	events := []int{5, 4, 9, 1, 2, 2, 10, 120, 100, 1, 3, 4}

	// first listener to be stopped after two of the events have been emitted
	expectedL1Events := events[:2]
	expectedL2Events := append([]int{}, events...)
	expectedL3Events := []int{120, 100}

	sort.Ints(expectedL1Events)
	sort.Ints(expectedL2Events)
	sort.Ints(expectedL3Events)

	done := make(chan bool)

	go func() {
		defer func() {
			l2.Stop()
			l3.Stop()
			done <- true
		}()
		for i, ev := range events {
			if i == 2 {
				l1.Stop()
			}
			em.Emit(ev)
		}
	}()

	var receivedL1Events, receivedL2Events, receivedL3Events []int

	var listenDone bool
	for !listenDone {
		select {
		case ev, ok := <-l1.Channel():
			if ok {
				receivedL1Events = append(receivedL1Events, ev)
			}
		case ev, ok := <-l2.Channel():
			if ok {
				receivedL2Events = append(receivedL2Events, ev)
			}
		case ev, ok := <-l3.Channel():
			if ok {
				receivedL3Events = append(receivedL3Events, ev)
			}
		case <-done:
			listenDone = true
		case <-time.After(250 * time.Millisecond):
			t.Fatal("test timed out")
		}
	}

	sort.Ints(receivedL1Events)
	sort.Ints(receivedL2Events)
	sort.Ints(receivedL3Events)

	if !reflect.DeepEqual(expectedL1Events, receivedL1Events) {
		t.Errorf("expected listener 1 events: %#v, got: %#v", expectedL1Events, receivedL1Events)
	}

	if !reflect.DeepEqual(expectedL2Events, receivedL2Events) {
		t.Errorf("expected listener 2 events: %#v, got: %#v", expectedL2Events, receivedL2Events)
	}

	if !reflect.DeepEqual(expectedL3Events, receivedL3Events) {
		t.Errorf("expected listener 3 events: %#v, got: %#v", expectedL3Events, receivedL3Events)
	}
}
