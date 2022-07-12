package event

import (
	"context"
)

type emitter[T any] struct {
	broadcast      chan T
	listeners      map[*listener[T]]bool
	addListener    chan *listener[T]
	removeListener chan *listener[T]
}

// NewEmitter creates a new event emitter.
// Event emitter can register new listener with NewListener.
func NewEmitter[T any](ctx context.Context) *emitter[T] {
	em := &emitter[T]{
		listeners:      make(map[*listener[T]]bool),
		broadcast:      make(chan T),
		addListener:    make(chan *listener[T]),
		removeListener: make(chan *listener[T]),
	}
	go em.run(ctx)
	return em
}

func (em *emitter[T]) run(ctx context.Context) {
	for {
		select {
		case event := <-em.broadcast:
			for l := range em.listeners {
				if l.match(event) {
					l.send <- event
				}
			}
		case l := <-em.addListener:
			em.listeners[l] = true
		case l := <-em.removeListener:
			if _, ok := em.listeners[l]; ok {
				delete(em.listeners, l)
				close(l.send)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (em emitter[T]) Emit(event T) {
	em.broadcast <- event
}

type listener[T any] struct {
	send  chan T
	match func(ev T) bool
	stop  func()
}

func (o *listener[T]) Stop() { o.stop() }

func (o *listener[T]) Channel() <-chan T { return o.send }

// listenerOpt applies an option to given listener
type listenerOpt[T any] func(l *listener[T])

// On adds a filter on the listener to only listen to matching event
func On[T any](match func(ev T) bool) listenerOpt[T] {
	return func(l *listener[T]) {
		l.match = match
	}
}

// NewListener creates a new listener with given options.
func (em *emitter[T]) NewListener(opts ...listenerOpt[T]) *listener[T] {
	l := &listener[T]{
		send: make(chan T),
		// By default, every event is listened to.
		match: func(ev T) bool { return true },
	}
	l.stop = func() { em.removeListener <- l }
	for _, opt := range opts {
		opt(l)
	}
	em.addListener <- l
	return l
}
