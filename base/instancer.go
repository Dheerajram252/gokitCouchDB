package base

import (
	"errors"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"gokitCouchDB/db"
	"sort"
	"sync"
	"sync/atomic"
)

type RoundRobin struct {
	databases      []db.DatabaseInterface
	c              uint64
	ch             chan sd.Event
	instancer      *consul.Instancer
	mutex          sync.RWMutex
	dbName         string
	dbType         string
	serviceGroup   string
	version        string
	requestCount   *prometheus.Counter
	errCount       *prometheus.Counter
	latency        *prometheus.Summary
	latencySeconds *prometheus.Histogram
	labels         []string
}

func NewRoundRobin(instancer *consul.Instancer, dbName string, dbType string, serviceGroup string, version string) *RoundRobin {
	labelNames := []string{"method"}
	constLabels := map[string]string{"dbName": dbName, "dbType": dbType, "version": version, "serviceGroup": serviceGroup}
	requestCount := prometheus.NewCounterFrom(
		stdPrometheus.CounterOpts{
			Name:        "request_Count",
			Subsystem:   "db",
			Help:        "Number of DB requests received",
			ConstLabels: constLabels,
		},
		labelNames,
	)

	errCount := prometheus.NewCounterFrom(
		stdPrometheus.CounterOpts{
			Name:        "err_count",
			Subsystem:   "db",
			Help:        "Number of DB errors",
			ConstLabels: constLabels,
		},
		labelNames,
	)

	requestLatency := prometheus.NewSummaryFrom(
		stdPrometheus.SummaryOpts{
			Name:        "request_latency_seconds",
			Subsystem:   "db",
			Help:        "Total duration of DB requests in request_latency_seconds",
			ConstLabels: constLabels,
		},
		labelNames,
	)

	requestLatencySeconds := prometheus.NewHistogramFrom(
		stdPrometheus.HistogramOpts{
			Subsystem:   "db",
			Name:        "request_latency",
			Help:        "Duration of DB request in seconds",
			ConstLabels: constLabels,
			Buckets:     []float64{.01, .025, .05, .1, .3, .6, 1},
		}, labelNames,
	)

	roundRobin := &RoundRobin{
		databases:      make([]db.DatabaseInterface, 0),
		c:              0,
		ch:             make(chan sd.Event),
		instancer:      instancer,
		dbName:         dbName,
		dbType:         dbType,
		serviceGroup:   serviceGroup,
		version:        version,
		requestCount:   requestCount,
		errCount:       errCount,
		latency:        requestLatency,
		latencySeconds: requestLatencySeconds,
		labels:         labelNames,
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
	databases := make([]db.DatabaseInterface, 0, len(insts))

	for _, i := range insts {
		database, err := db.NewClient(r.dbType, i, r.dbName)
		if err != nil {
			continue
		}
		database = NewDBInstrumentingService(r.labels, r.requestCount, r.errCount, r.latency, r.latencySeconds)(database)
		databases = append(databases, database)
	}
	r.databases = databases
}

func (r *RoundRobin) DB() (db.DatabaseInterface, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.databases) <= 0 {
		return db.DB{}, errors.New("no databases available")
	}

	old := atomic.AddUint64(&r.c, 1) - 1
	idx := old % uint64(len(r.databases))
	return r.databases[idx], nil
}
