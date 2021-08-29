package registry

import (
	"sort"
	"sync"
	"time"
)

type GeeRegistry struct {
	timeout time.Duration // server becomes unavailable after timeout
	mu      sync.Mutex    // protect following
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_geerpc_/registry"
	defaultTimeout = 5 * time.Second
)

func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultGeeRegistry = New(defaultTimeout)

func (r *GeeRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	s, ok := r.servers[addr]
	if ok {
		s.start = now
	} else {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: now,
		}
	}
}

func (r *GeeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			// not timeout yet
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}

	sort.Strings(alive)
	return alive
}
