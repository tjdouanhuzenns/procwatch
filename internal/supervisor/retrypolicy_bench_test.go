package supervisor

import (
	"testing"
	"time"
)

func BenchmarkRetryPolicy_AllowUnderLimit(b *testing.B) {
	m := NewRetryPolicyManager()
	_ = m.SetPolicy("svc", RetryPolicy{
		MaxAttempts: 1000,
		Window:      time.Minute,
		Cooldown:    time.Second,
	})
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Allow("svc", now)
	}
}

func BenchmarkRetryPolicy_RecordAndAllow(b *testing.B) {
	m := NewRetryPolicyManager()
	_ = m.SetPolicy("svc", RetryPolicy{
		MaxAttempts: 1000,
		Window:      time.Minute,
		Cooldown:    time.Second,
	})
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Record("svc", now)
		m.Allow("svc", now)
	}
}

func BenchmarkRetryPolicy_Parallel(b *testing.B) {
	m := NewRetryPolicyManager()
	_ = m.SetPolicy("svc", RetryPolicy{
		MaxAttempts: 1000,
		Window:      time.Minute,
		Cooldown:    time.Second,
	})
	now := time.Now()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Record("svc", now)
			m.Allow("svc", now)
		}
	})
}
