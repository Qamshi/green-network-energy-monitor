package main

import (
	"fmt"
	"net/http"

	"github.com/freeconf/restconf"
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/source"
)

func main() {
	fmt.Println("=======================================================")
	fmt.Println("  Program C — ISAC Utilization YANG RESTCONF Server")
	fmt.Println("  Draft: draft-jadoon-green-isac-utilization-03")
	fmt.Println("  Port : 8883")
	fmt.Println("=======================================================")

	ypath := source.Any(
		source.Dir("./yang"),
		restconf.InternalYPath,
	)

	dev := device.New(ypath)

	// "isac-utilization" must match module name in isac-utilization.yang
	if err := dev.Add("isac-utilization", ISACProvider()); err != nil {
		fmt.Printf("FATAL: Could not add module: %v\n", err)
		return
	}

	s := restconf.NewServer(dev)

	fmt.Println("----------------------------------------------")
	fmt.Println("RESTCONF Server running on http://localhost:8883")
	fmt.Println("Endpoint: http://localhost:8883/restconf/data/isac-utilization:isac-utilization")
	fmt.Println("----------------------------------------------")

	if err := http.ListenAndServe(":8883", s); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

