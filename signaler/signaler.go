package signaler

import (
	"sync"

	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("broadcaster")
)

type Signaler struct {
	lock      sync.Mutex
	listeners []chan struct{}
}

func New() *Signaler {
	return &Signaler{}
}

func (s *Signaler) Listen() <-chan struct{} {
	c := make(chan struct{}, 1)
	s.lock.Lock()
	s.listeners = append(s.listeners, c)
	s.lock.Unlock()
	return c
}

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
			return
		}
	}
}

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
