package ratelim

import (
	"fmt"
	"sync"
)

type RateLim struct {
	c chan struct{}

	lock   sync.Mutex
	errors []string
}

func New(limit int) (*RateLim, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit should be positive")
	}

	return &RateLim{c: make(chan struct{}, limit)}, nil
}

func (rl *RateLim) Exec(f func() error) {
	rl.c <- struct{}{}
	defer func() { <-rl.c }()
	err := f()
	if err != nil {
		rl.lock.Lock()
		rl.errors = append(rl.errors, err.Error())
		rl.lock.Unlock()
	}
}

func (rl *RateLim) Wait() []string {
	for i := 0; i < cap(rl.c); i++ {
		rl.c <- struct{}{}
	}
	return rl.errors
}
