package metrics

import (
	"sort"
	"time"
)

type Metrics struct {
	Throughput float64
	ErrorRate  float64
	Latencies  []time.Duration
	CPU        float64
	Memory     float64
	Network    float64
}

func (m *Metrics) Aggregate(other Metrics) {
	m.Throughput += other.Throughput
	m.ErrorRate += other.ErrorRate
	m.Latencies = append(m.Latencies, other.Latencies...)
}

func (m *Metrics) CalculateLatencyPercentiles() struct {
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
} {
	if len(m.Latencies) == 0 {
		return struct {
			P90 time.Duration
			P95 time.Duration
			P99 time.Duration
		}{}
	}
	sort.Slice(m.Latencies, func(i, j int) bool {
		return m.Latencies[i] < m.Latencies[j]
	})

	p90Index := int(float64(len(m.Latencies)) * 0.9)
	p95Index := int(float64(len(m.Latencies)) * 0.95)
	p99Index := int(float64(len(m.Latencies)) * 0.99)

	return struct {
		P90 time.Duration
		P95 time.Duration
		P99 time.Duration
	}{
		P90: m.Latencies[p90Index],
		P95: m.Latencies[p95Index],
		P99: m.Latencies[p99Index],
	}
}
