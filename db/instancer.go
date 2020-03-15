package db

import (
	"errors"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"sort"
	"sync"
	"sync/atomic"
)

type RoundRobin struct {
	databases []DB
	c         uint64
	ch        chan sd.Event
	instancer *consul.Instancer
	mutex     sync.RWMutex
	dbName    string
	dbType    string
}

func NewRoundRobin(instancer *consul.Instancer, dbName string, dbType string) *RoundRobin {
	roundRobin := &RoundRobin{
		databases: make([]DB, 0),
		c:         0,
		ch:        make(chan sd.Event),
		instancer: instancer,
		dbName:    dbName,
		dbType:    dbType,
	}
	go roundRobin.receive()
	instancer.Register(roundRobin.ch)
	return roundRobin
}

func (r *RoundRobin) receive() {
	for event := range r.ch {
		r.Update(event)
	}
}

func (r *RoundRobin) Update(ev sd.Event) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if ev.Err != nil {
		return
	}
	insts := ev.Instances
	sort.Strings(insts)
	databases := make([]DB, 0, len(insts))

	for _, i := range insts {
		database, err := NewClient(r.dbType, i, r.dbName)
		if err != nil {
			continue
		}
		databases = append(databases, database)
	}
	r.databases = databases
}

func (r *RoundRobin) DB() (DB, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.databases) <= 0 {
		return DB{}, errors.New("no databases available")
	}

	old := atomic.AddUint64(&r.c, 1) - 1
	idx := old % uint64(len(r.databases))
	return r.databases[idx], nil
}
