package main

import (
	"fmt"
	"net/http"
	"github.com/freeconf/restconf"
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/source"
)

func main() {
	fc.DebugLog(true)

	// ===========>> Point to your YANG directory
	ypath := source.Any(
		source.Dir("../../yang"),
		restconf.InternalYPath,
	)

	dev := device.New(ypath)

	// ===========>> Wrap the provider so the server finds the "devices" container first
	err := dev.Add("sonoff-energy", &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "devices" {
				return DevicesProvider(), nil
			}
			return nil, nil
		},
	})

	if err != nil {
		fmt.Printf("FATAL: Could not add module: %v\n", err)
		return
	}

	// ===========>> Initialize Server
	s := restconf.NewServer(dev)

	fmt.Println("----------------------------------------------")
	fmt.Println("RESTCONF Server starting on http://localhost:8080")
	fmt.Println("Testing URL: http://localhost:8080/restconf/data/sonoff-energy:devices")
	fmt.Println("----------------------------------------------")

	// 4. Listen and Serve
	listenErr := http.ListenAndServe(":8080", s)
	if listenErr != nil {
		fmt.Printf("Error starting server: %s\n", listenErr)
	}
}