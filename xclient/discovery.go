package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	Random SelectMode = iota
	RoundRobin
)

type Discovery interface {
	Refresh() error                      // refresh server list from remote registry
	Update(servers []string) error       // set server list
	Get(mode SelectMode) (string, error) // get next server by SelectMode
	GetAll() ([]string, error)           // get the whole server list
}

type MultiServersDiscovery struct {
	r       *rand.Rand   // generate random number
	mu      sync.RWMutex // protect following
	servers []string
	index   int // the selected positions for round robin mode
}

func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// generate random index
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServersDiscovery)(nil)

func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.servers = servers
	return nil
}

func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available server")
	}

	switch mode {
	case Random:
		// select random server from list
		return d.servers[d.r.Intn(n)], nil
	case RoundRobin:
		s := d.servers[d.index%n] // ensure index is inside the bound
		d.index = (d.index+1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: unsupported select mode " + string(mode))
	}
	return "", nil
}

func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	/// return a copy of server list
	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
