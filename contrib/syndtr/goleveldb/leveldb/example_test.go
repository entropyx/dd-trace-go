package leveldb_test

import (
	"context"

	leveldbtrace "github.com/entropyx/dd-trace-go/contrib/syndtr/goleveldb/leveldb"
	"github.com/entropyx/dd-trace-go/ddtrace/tracer"
)

func Example() {
	db, _ := leveldbtrace.OpenFile("/tmp/example.leveldb", nil)

	// Create a root span, giving name, server and resource.
	_, ctx := tracer.StartSpanFromContext(context.Background(), "my-query",
		tracer.ServiceName("my-db"),
		tracer.ResourceName("initial-access"),
	)

	// use WithContext to associate the span with the parent
	db.WithContext(ctx).
		// calls will be traced
		Put([]byte("key"), []byte("value"), nil)
}
