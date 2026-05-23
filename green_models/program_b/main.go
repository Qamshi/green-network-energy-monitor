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
	fmt.Println("  Program B — SmartPDU YANG RESTCONF Server")
	fmt.Println("  Draft: draft-ahc-green-smartpdu-yang-00")
	fmt.Println("  Port : 8882")
	fmt.Println("=======================================================")

	ypath := source.Any(
		source.Dir("./yang"),
		restconf.InternalYPath,
	)

	dev := device.New(ypath)

	// "smartpdu" must match module name in smartpdu.yang
	if err := dev.Add("smartpdu", SmartPDUProvider()); err != nil {
		fmt.Printf("FATAL: Could not add module: %v\n", err)
		return
	}

	s := restconf.NewServer(dev)

	fmt.Println("----------------------------------------------")
	fmt.Println("RESTCONF Server running on http://localhost:8882")
	fmt.Println("Endpoint: http://localhost:8882/restconf/data/smartpdu:pdu-system")
	fmt.Println("----------------------------------------------")

	if err := http.ListenAndServe(":8882", s); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

