package server

import "sync"

type sseMessage struct {
	event string
	data  string
}

// broadcaster fans out SSE messages to all subscribed clients.
// Slow clients are dropped rather than blocking the publisher.
type broadcaster struct {
	mu     sync.Mutex
	subs   map[chan sseMessage]struct{}
	closed bool
}

func newBroadcaster() *broadcaster {
	return &broadcaster{subs: make(map[chan sseMessage]struct{})}
}

func (b *broadcaster) subscribe() chan sseMessage {
	ch := make(chan sseMessage, 4)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		close(ch)
		return ch
	}
	b.subs[ch] = struct{}{}
	return ch
}

func (b *broadcaster) unsubscribe(ch chan sseMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
}

func (b *broadcaster) publish(msg sseMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default:
			// Drop if the subscriber is not keeping up.
		}
	}
}

func (b *broadcaster) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for ch := range b.subs {
		delete(b.subs, ch)
		close(ch)
	}
}
