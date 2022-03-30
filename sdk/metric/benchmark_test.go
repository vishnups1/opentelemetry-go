package metric

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/reader"
)

type testExporter struct {
	producer reader.Producer
}

func (t *testExporter) Register(producer reader.Producer) {
	t.producer = producer
}

func (*testExporter) Flush(context.Context) error { return nil }

func (*testExporter) Shutdown(context.Context) error { return nil }

func BenchmarkCounterAddNoAttrs(b *testing.B) {
	ctx := context.Background()
	exp := &testExporter{}
	rdr := reader.New(exp)
	provider := New(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1)
	}
}

// Benchmark prints 3 allocs per Add():
//  1. new []attribute.KeyValue for the list of attributes
//  2. interface{} wrapper around attribute.Set
//  3. an attribute array (map key)
func BenchmarkCounterAddOneAttr(b *testing.B) {
	ctx := context.Background()
	exp := &testExporter{}
	rdr := reader.New(exp)
	provider := New(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.String("K", "V"))
	}
}

// Benchmark prints 11 allocs per Add(), I see 10 in the profile:
//  1. new []attribute.KeyValue for the list of attributes
//  2. an attribute.Sortable (acquireRecord)
//  3. the attribute.Set underlying array
//  4. interface{} wrapper around attribute.Set value
//  5. internal to sync.Map
//  6. internal sync.Map
//  7. new syncstate.record
//  8. new viewstate.syncAccumulator
//  9. an attribute.Sortable (findOutput)
// 10. an output Aggregator
func BenchmarkCounterAddManyAttrs(b *testing.B) {
	ctx := context.Background()
	exp := &testExporter{}
	rdr := reader.New(exp)
	provider := New(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("K", i))
	}
}

func BenchmarkCounterCollectOneAttr(b *testing.B) {
	ctx := context.Background()
	exp := &testExporter{}
	rdr := reader.New(exp)
	provider := New(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("K", 1))
		_ = exp.producer.Produce()
	}
}

func BenchmarkCounterCollectTenAttrs(b *testing.B) {
	ctx := context.Background()
	exp := &testExporter{}
	rdr := reader.New(exp)
	provider := New(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			cntr.Add(ctx, 1, attribute.Int("K", j))
		}
		_ = exp.producer.Produce()
	}
}
