package main

import (
	"flag"
	"fmt"

	"druk/pkg/ui"
)

func main() {
	endpoint := flag.String("endpoint", "", "API endpoint to test")
	flag.Parse()

	if *endpoint == "" {
		fmt.Println("Please provide an endpoint using the -endpoint flag")
		return
	}

	p := ui.NewProgram(*endpoint)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
