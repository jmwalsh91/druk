package main

import (
	"druk/pkg/ui"
	"flag"
	"fmt"
	"time"
)

func main() {
	endpoint := flag.String("endpoint", "", "API endpoint to test")
	duration := flag.Duration("duration", 5*time.Second, "Duration of the load test")
	concurrency := flag.Int("concurrency", 4, "Number of concurrent requests")
	flag.Parse()

	if *endpoint == "" {
		fmt.Println("Please provide an endpoint using the -endpoint flag")
		return
	}

	p := ui.NewProgram(*endpoint, *duration, *concurrency)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
