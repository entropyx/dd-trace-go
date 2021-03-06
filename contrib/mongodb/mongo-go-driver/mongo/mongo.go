// Package mongo provides functions to trace the mongodb/mongo-go-driver package (https://go.mongodb.org/mongo-driver).
//
// `NewMonitor` will return an event.CommandMonitor which is used to trace requests.
package mongo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/entropyx/dd-trace-go/ddtrace"
	"github.com/entropyx/dd-trace-go/ddtrace/ext"
	"github.com/entropyx/dd-trace-go/ddtrace/tracer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
)

type spanKey struct {
	ConnectionID string
	RequestID    int64
}

type monitor struct {
	sync.Mutex
	spans map[spanKey]ddtrace.Span
}

func (m *monitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	hostname, port := peerInfo(evt)
	b, _ := bson.MarshalExtJSON(evt.Command, false, false)
	span, _ := tracer.StartSpanFromContext(ctx, "mongodb.query",
		tracer.ServiceName("mongo"),
		tracer.ResourceName("mongo."+evt.CommandName),
		tracer.Tag(ext.DBInstance, evt.DatabaseName),
		tracer.Tag(ext.DBStatement, string(b)),
		tracer.Tag(ext.DBType, "mongo"),
		tracer.Tag(ext.PeerHostname, hostname),
		tracer.Tag(ext.PeerPort, port),
	)
	key := spanKey{
		ConnectionID: evt.ConnectionID,
		RequestID:    evt.RequestID,
	}
	m.Lock()
	m.spans[key] = span
	m.Unlock()
}

func (m *monitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	m.Finished(&evt.CommandFinishedEvent, nil)
}

func (m *monitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	m.Finished(&evt.CommandFinishedEvent, fmt.Errorf("%s", evt.Failure))
}

func (m *monitor) Finished(evt *event.CommandFinishedEvent, err error) {
	key := spanKey{
		ConnectionID: evt.ConnectionID,
		RequestID:    evt.RequestID,
	}
	m.Lock()
	span, ok := m.spans[key]
	if ok {
		delete(m.spans, key)
	}
	m.Unlock()
	if !ok {
		return
	}
	span.Finish(tracer.WithError(err))
}

// NewMonitor creates a new mongodb event CommandMonitor.
func NewMonitor() *event.CommandMonitor {
	m := &monitor{
		spans: make(map[spanKey]ddtrace.Span),
	}
	return &event.CommandMonitor{
		Started:   m.Started,
		Succeeded: m.Succeeded,
		Failed:    m.Failed,
	}
}

func peerInfo(evt *event.CommandStartedEvent) (hostname, port string) {
	hostname = evt.ConnectionID
	port = "27017"
	if idx := strings.IndexByte(hostname, '['); idx >= 0 {
		hostname = hostname[:idx]
	}
	if idx := strings.IndexByte(hostname, ':'); idx >= 0 {
		port = hostname[idx+1:]
		hostname = hostname[:idx]
	}
	return hostname, port
}
