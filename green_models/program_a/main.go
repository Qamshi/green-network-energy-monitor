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
	fmt.Println("  Program A — Power & Energy YANG RESTCONF Server")
	fmt.Println("  Draft: draft-bcmj-green-power-and-energy-yang-04")
	fmt.Println("  Port : 8881")
	fmt.Println("=======================================================")

	// Same pattern as your existing main.go:
	// source.Dir points to our yang/ folder
	// restconf.InternalYPath adds FreeCONF's built-in YANG files
	ypath := source.Any(
		source.Dir("./yang"),
		restconf.InternalYPath,
	)

	dev := device.New(ypath)

	// Register our module — "power-energy" must match module name in power-energy.yang
	if err := dev.Add("power-energy", PowerEnergyProvider()); err != nil {
		fmt.Printf("FATAL: Could not add module: %v\n", err)
		return
	}

	// Start RESTCONF server — same as your existing project
	s := restconf.NewServer(dev)

	fmt.Println("----------------------------------------------")
	fmt.Println("RESTCONF Server running on http://localhost:8881")
	fmt.Println("Endpoint: http://localhost:8881/restconf/data/power-energy:energy-objects")
	fmt.Println("----------------------------------------------")

	if err := http.ListenAndServe(":8881", s); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

