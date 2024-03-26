package loadtest

import (
	"log"
	"net/http"
	"sync"
	"time"

	"druk/pkg/metrics"
)

func Run(endpoint string, duration time.Duration, concurrency int, progressCh chan<- float64) (metrics.Metrics, error) {
	client := &http.Client{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	start := time.Now()

	var totalRequests int64
	var totalLatency time.Duration
	var errorCount int64
	statusCodes := make(map[int]int)
	errors := make(map[string]int)
	latencies := make([]time.Duration, 0)

	log.Printf("Starting load test with endpoint: %s, duration: %s, concurrency: %d", endpoint, duration, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Since(start) < duration {
				reqStart := time.Now()
				resp, err := client.Get(endpoint)
				latency := time.Since(reqStart)

				mu.Lock()
				totalRequests++
				totalLatency += latency
				statusCodes[resp.StatusCode]++
				latencies = append(latencies, latency)
				if err != nil {
					errorCount++
					errors[err.Error()]++
				} else if resp.StatusCode >= 400 {
					errorCount++
					errors[resp.Status]++
				}
				mu.Unlock()

				if err == nil {
					resp.Body.Close()
				}
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				progress := float64(time.Since(start)) / float64(duration)
				progressCh <- progress
			case <-time.After(duration):
				close(progressCh)
				return
			}
		}
	}()

	wg.Wait()

	result := metrics.Metrics{
		Throughput:    float64(totalRequests) / duration.Seconds(),
		ErrorRate:     float64(errorCount) / float64(totalRequests) * 100,
		Latencies:     latencies,
		Duration:      duration,
		StatusCodes:   statusCodes,
		Errors:        errors,
		TotalRequests: int(totalRequests),
	}

	log.Printf("Metrics: %+v", result)
	return result, nil
}
