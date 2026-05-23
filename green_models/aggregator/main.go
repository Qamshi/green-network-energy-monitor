package main

import (
	"fmt"
	"net/http"

	"github.com/freeconf/restconf"
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/source"
)

// corsMiddleware adds CORS headers so the browser dashboard can fetch from this server
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("=======================================================")
	fmt.Println("  Aggregator — Unified Energy RESTCONF Server")
	fmt.Println("  Sources: Program A (8881) + B (8882) + C (8883) + Sonoff (8080)")
	fmt.Println("  Port   : 8884")
	fmt.Println("=======================================================")

	ypath := source.Any(
		source.Dir("./yang"),
		restconf.InternalYPath,
	)

	dev := device.New(ypath)

	err := dev.Add("aggregator", AggregatorProvider())
	if err != nil {
		fmt.Printf("FATAL: Could not add module: %v\n", err)
		return
	}

	s := restconf.NewServer(dev)

	fmt.Println("----------------------------------------------")
	fmt.Println("Aggregator running on http://localhost:8884")
	fmt.Println("Endpoint: http://localhost:8884/restconf/data/aggregator:aggregated-energy")
	fmt.Println("Dashboard: open dashboard.html in your browser")
	fmt.Println("----------------------------------------------")

	listenErr := http.ListenAndServe(":8884", corsMiddleware(s))
	if listenErr != nil {
		fmt.Printf("Error starting server: %s\n", listenErr)
	}
}
