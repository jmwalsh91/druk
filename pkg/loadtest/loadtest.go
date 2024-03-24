package loadtest

import (
	"net/http"
	"time"

	"druk/pkg/metrics"
)

func Run(endpoint string, duration time.Duration, concurrency int) (metrics.Metrics, error) {
	client := &http.Client{}
	done := make(chan struct{})
	metricsChan := make(chan metrics.Metrics)

	for i := 0; i < concurrency; i++ {
		go func() {
			start := time.Now()
			var collectedMetrics metrics.Metrics
			for time.Since(start) < duration {
				reqStart := time.Now()
				resp, err := client.Get(endpoint)
				latency := time.Since(reqStart)
				if err != nil {
					collectedMetrics.ErrorRate++
				} else {
					collectedMetrics.Throughput++
					collectedMetrics.Latencies = append(collectedMetrics.Latencies, latency)
				}
				resp.Body.Close()
			}
			metricsChan <- collectedMetrics
		}()
	}

	go func() {
		time.Sleep(duration)
		close(done)
	}()

	var finalMetrics metrics.Metrics
	for i := 0; i < concurrency; i++ {
		select {
		case m := <-metricsChan:
			finalMetrics.Aggregate(m)
		case <-done:
			return finalMetrics, nil
		}
	}

	return finalMetrics, nil
}
