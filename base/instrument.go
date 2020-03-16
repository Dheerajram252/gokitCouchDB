package base

import (
	"context"
	"github.com/go-kit/kit/metrics"
	"gokitCouchDB/db"
	"net/http"
	"strconv"
	"time"
)

type contextKey string

const contextStartTime = contextKey("startTime")

type InstrumentingMiddleware func(Service) Service

type instrumentingService struct {
	instrumenter instrumenter
	next         Service
}

type instrumenter struct {
	labelNames            []string
	requestCount          metrics.Counter
	errCount              metrics.Counter
	requestLatencySummary metrics.Histogram
	requestLatency        metrics.Histogram
}

func NewInstrumentingService(lablelNames []string, counter metrics.Counter, errCounter metrics.Counter, latencySummary metrics.Histogram, histogram metrics.Histogram) InstrumentingMiddleware {
	return func(next Service) Service {
		return instrumentingService{
			instrumenter: instrumenter{
				labelNames:            lablelNames,
				requestCount:          counter,
				errCount:              errCounter,
				requestLatencySummary: latencySummary,
				requestLatency:        histogram,
			},
			next: next,
		}
	}
}

func (s instrumenter) instrument(begin time.Time, methodName string, err error) {
	if len(s.labelNames) > 0 {
		s.requestCount.With(s.labelNames[0], methodName).Add(1)
		s.requestLatencySummary.With(s.labelNames[0], methodName).Observe(time.Since(begin).Seconds())
		s.requestLatency.With(s.labelNames[0], methodName).Observe(time.Since(begin).Seconds())
		if err != nil {
			s.errCount.With(s.labelNames[0], methodName).Add(1)
		}
	}
}

func (s instrumentingService) Check(ctx context.Context) (p bool, err error) {
	defer func(begin time.Time, err error) {
		s.instrumenter.instrument(begin, "Check", err)
	}(time.Now(), err)
	return s.next.Check(ctx)
}

func (s instrumentingService) GetDocument(ctx context.Context, req string) (response interface{}, err error) {
	defer func(begin time.Time, err error) {
		s.instrumenter.instrument(begin, "GetDocument", err)
	}(time.Now(), err)
	return s.next.GetDocument(ctx, req)
}

func (s instrumentingService) PutDocument(ctx context.Context) (p bool, err error) {
	defer func(begin time.Time, err error) {
		s.instrumenter.instrument(begin, "PutDocument", err)
	}(time.Now(), err)
	return s.next.PutDocument(ctx)
}

type TransportServerFinalizerInstrument interface {
	TransportServerFinalizer(context.Context, int, *http.Request)
}

type transportServerFinalizerInstrument struct {
	labelNames     []string
	requestLatency metrics.Histogram
}

func NewTransportServerFinalizerInstrument(labelNames []string, requestLatency metrics.Histogram) TransportServerFinalizerInstrument {
	return &transportServerFinalizerInstrument{
		labelNames:     labelNames,
		requestLatency: requestLatency,
	}
}

func startTime(ctx context.Context) time.Time {
	value, _ := ctx.Value(contextStartTime).(time.Time)
	return value
}

func (t transportServerFinalizerInstrument) TransportServerFinalizer(ctx context.Context, code int, r *http.Request) {
	if len(t.labelNames) > 1 {
		t.requestLatency.With(t.labelNames[0], r.URL.RequestURI(), t.labelNames[1], strconv.Itoa(code)).Observe(time.Since(startTime(ctx)).Seconds())
	}
}

type DBInstrumentingMiddleware func(db.DatabaseInterface) db.DatabaseInterface

type dbInstrumentingService struct {
	instrumenter instrumenter
	next         db.DatabaseInterface
}

func NewDBInstrumentingService(labelNames []string, counter metrics.Counter, errCounter metrics.Counter, latencySummary metrics.Histogram, histogram metrics.Histogram) DBInstrumentingMiddleware {
	return func(next db.DatabaseInterface) db.DatabaseInterface {
		return dbInstrumentingService{
			instrumenter: instrumenter{
				labelNames:            labelNames,
				requestCount:          counter,
				errCount:              errCounter,
				requestLatency:        histogram,
				requestLatencySummary: latencySummary,
			},
			next: next,
		}
	}
}

func (db dbInstrumentingService) GetDocument(ctx context.Context, docID string) (res interface{}, err error) {
	defer func(begin time.Time, err error) {
		db.instrumenter.instrument(begin, "GetDocument", err)
	}(time.Now(), err)
	return db.next.GetDocument(ctx, docID)
}

func (db dbInstrumentingService) PutDocument(ctx context.Context) (flag bool, err error) {
	defer func(begin time.Time, err error) {
		db.instrumenter.instrument(begin, "GetDocument", err)
	}(time.Now(), err)
	return db.next.PutDocument(ctx)
}
