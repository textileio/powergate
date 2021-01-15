package ratelim

import (
	"fmt"
	"sync"
)

// RateLim allows to logically cap the maximum number of functions.
// It returns a slice of error strings that happened during execution.
type RateLim struct {
	c chan struct{}

	lock   sync.Mutex
	errors []string
}

// New returns a new RateLim.
func New(limit int) (*RateLim, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit should be positive")
	}

	return &RateLim{c: make(chan struct{}, limit)}, nil
}

// Exec executes f when a slot is available within the defined
// maximum limit.
func (rl *RateLim) Exec(f func() error) {
	rl.c <- struct{}{}
	go func() {
		defer func() { <-rl.c }()
		err := f()
		if err != nil {
			rl.lock.Lock()
			rl.errors = append(rl.errors, err.Error())
			rl.lock.Unlock()
		}
	}()
}

// Wait will block until all executing functions finish, and return
// all error strings that happened during executions. RateLim can't
// be reused after this call.
func (rl *RateLim) Wait() []string {
	for i := 0; i < cap(rl.c); i++ {
		rl.c <- struct{}{}
	}
	return rl.errors
}
