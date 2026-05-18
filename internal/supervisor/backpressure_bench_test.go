package supervisor

import (
	"context"
	"testing"
)

func BenchmarkBackpressure_AcquireRelease(b *testing.B) {
	bp := NewBackpressure(BackpressureConfig{
		MaxPending:  100,
		WaitTimeout: 0,
	})
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if bp.Acquire(ctx) {
			bp.Release()
		}
	}
}

func BenchmarkBackpressure_Pending(b *testing.B) {
	bp := NewBackpressure(DefaultBackpressureConfig())
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		bp.Acquire(ctx)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bp.Pending()
	}
}

func BenchmarkBackpressure_Parallel(b *testing.B) {
	bp := NewBackpressure(BackpressureConfig{
		MaxPending:  1000,
		WaitTimeout: 0,
	})
	ctx := context.Background()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if bp.Acquire(ctx) {
				bp.Release()
			}
		}
	})
}
