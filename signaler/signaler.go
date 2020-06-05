package signaler

import (
	"sync"

	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("broadcaster")
)

// Signaler allows subscribing to a singnaling hub.
type Signaler struct {
	lock      sync.Mutex
	listeners []chan struct{}
}

// New returns a new Signaler.
func New() *Signaler {
	return &Signaler{}
}

// Listen returns a new channel signaler.
func (s *Signaler) Listen() <-chan struct{} {
	c := make(chan struct{}, 1)
	s.lock.Lock()
	s.listeners = append(s.listeners, c)
	s.lock.Unlock()
	return c
}

// Unregister unregisters a channel signaler from the hub.
func (s *Signaler) Unregister(c chan struct{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i := range s.listeners {
		if s.listeners[i] == c {
			if len(s.listeners) == 1 {
				s.listeners = nil
				return
			}
			s.listeners[i] = s.listeners[len(s.listeners)-1]
			s.listeners = s.listeners[:len(s.listeners)-1]
			close(s.listeners[i])
			return
		}
	}
}

// Signal triggers a new notification to all listeners.
func (s *Signaler) Signal() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, c := range s.listeners {
		select {
		case c <- struct{}{}:
		default:
			log.Warn("dropping signal on blocked listener")
		}
	}
}

// Close closes the Signaler. Any channel that wasn't explicitly unregistered,
// is closed.
func (s *Signaler) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, c := range s.listeners {
		close(c)
	}
}
