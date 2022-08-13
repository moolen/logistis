package locker

import "sync"

type Locker struct {
	mu *sync.RWMutex

	lockMessages map[string]string
}

func New() *Locker {
	return &Locker{
		mu:           &sync.RWMutex{},
		lockMessages: make(map[string]string),
	}
}

func (l *Locker) HandleLock() {

}
