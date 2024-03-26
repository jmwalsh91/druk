package metrics

import (
	"sort"
	"time"
)

type Metrics struct {
	Throughput        float64
	ErrorRate         float64
	Latencies         []time.Duration
	Duration          time.Duration
	AvgLatency        time.Duration
	LatencyP90        time.Duration
	LatencyP95        time.Duration
	LatencyP99        time.Duration
	TotalRequests     int
	LatencyData       []float64
	ThroughputData    []float64
	StatusCodes       map[int]int
	Errors            map[string]int
	RequestsPerSecond float64
	DataTransferred   float64
	OpenFiles         int
	MaxOpenFiles      int
}

func (m *Metrics) Aggregate(other Metrics) {
	m.Throughput += other.Throughput
	m.ErrorRate += other.ErrorRate
	m.Latencies = append(m.Latencies, other.Latencies...)
	m.TotalRequests += other.TotalRequests
}

func (m *Metrics) CalculateStatistics() {
	m.TotalRequests = int(m.Throughput * m.Duration.Seconds())
	m.RequestsPerSecond = float64(m.TotalRequests) / m.Duration.Seconds()

	if len(m.Latencies) > 0 {
		sort.Slice(m.Latencies, func(i, j int) bool {
			return m.Latencies[i] < m.Latencies[j]
		})

		var totalLatency time.Duration
		for _, latency := range m.Latencies {
			totalLatency += latency
		}
		m.AvgLatency = totalLatency / time.Duration(len(m.Latencies))

		p90Index := int(float64(len(m.Latencies)) * 0.9)
		p95Index := int(float64(len(m.Latencies)) * 0.95)
		p99Index := int(float64(len(m.Latencies)) * 0.99)

		m.LatencyP90 = m.Latencies[p90Index]
		m.LatencyP95 = m.Latencies[p95Index]
		m.LatencyP99 = m.Latencies[p99Index]

		m.LatencyData = make([]float64, len(m.Latencies))
		for i, latency := range m.Latencies {
			m.LatencyData[i] = float64(latency) / float64(time.Millisecond)
		}
	}

	m.ErrorRate = (m.ErrorRate / float64(m.TotalRequests)) * 100

	m.ThroughputData = make([]float64, int(m.Duration.Seconds()))
	for i := 0; i < len(m.ThroughputData); i++ {
		secondStartTime := time.Duration(i) * time.Second
		secondEndTime := time.Duration(i+1) * time.Second
		var requestsInSecond int
		for _, latency := range m.Latencies {
			if latency >= secondStartTime && latency < secondEndTime {
				requestsInSecond++
			}
		}
		m.ThroughputData[i] = float64(requestsInSecond)
	}
}
