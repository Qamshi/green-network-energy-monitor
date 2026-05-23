package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/val"
)

// ---------------------------------------------------------------------------
// Data structures matching smartpdu.yang
// ---------------------------------------------------------------------------

type InputLine struct {
	ID        string  `json:"id"`
	Voltage   float64 `json:"voltage"`
	Current   float64 `json:"current"`
	Frequency float64 `json:"frequency"`
	Power     float64 `json:"power"`
	Energy    float64 `json:"energy"`
	Status    string  `json:"status"`
}

type Outlet struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Power       float64 `json:"power"`
	Energy      float64 `json:"energy"`
	PowerFactor float64 `json:"power-factor"`
	Status      string  `json:"status"`
}

type Sensor struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Status string  `json:"status"`
}

type PDUSystem struct {
	Manufacturer    string  `json:"manufacturer"`
	Model           string  `json:"model"`
	SerialNumber    string  `json:"serial-number"`
	FirmwareVersion string  `json:"firmware-version"`
	Uptime          uint64  `json:"uptime"`
	Temperature     float64 `json:"temperature"`
	Humidity        float64 `json:"humidity"`
	InputLines      struct {
		Line []InputLine `json:"line"`
	} `json:"input-lines"`
	Outlets struct {
		Outlet []Outlet `json:"outlet"`
	} `json:"outlets"`
	Sensors struct {
		Sensor []Sensor `json:"sensor"`
	} `json:"sensors"`
}

type PDUDataFile struct {
	PDUSystem PDUSystem `json:"pdu-common:pdu-system"`
}

func loadDatastore() (PDUSystem, error) {
	var empty PDUSystem
	f, err := os.Open("./data.json")
	if err != nil {
		fmt.Println("ERROR: Could not open data.json")
		fmt.Println("Tip: Run 'python generator.py' first.")
		return empty, err
	}
	defer f.Close()

	var df PDUDataFile
	if err := json.NewDecoder(f).Decode(&df); err != nil {
		fmt.Printf("ERROR: Failed to decode data.json: %v\n", err)
		return empty, err
	}
	fmt.Printf("INFO: Loaded PDU data — %d outlets, %d sensors\n",
		len(df.PDUSystem.Outlets.Outlet), len(df.PDUSystem.Sensors.Sensor))
	return df.PDUSystem, nil
}

// ---------------------------------------------------------------------------
// SmartPDUProvider — root FreeCONF node
//
// YANG tree:
//   module smartpdu
//     container pdu-system              <-- OnChild at root
//       leaf manufacturer, model, ...
//       container input-lines           <-- OnChild inside pdu-system
//         list line                     <-- OnNext inside input-lines
//       container outlets               <-- OnChild inside pdu-system
//         list outlet                   <-- OnNext inside outlets
//       container sensors               <-- OnChild inside pdu-system
//         list sensor                   <-- OnNext inside sensors
// ---------------------------------------------------------------------------

func SmartPDUProvider() node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "pdu-system" {
				pdu, err := loadDatastore()
				if err != nil {
					return nil, err
				}
				return buildPDUSystemNode(pdu), nil
			}
			return nil, nil
		},
	}
}

func buildPDUSystemNode(pdu PDUSystem) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "manufacturer":     raw = pdu.Manufacturer
			case "model":            raw = pdu.Model
			case "serial-number":    raw = pdu.SerialNumber
			case "firmware-version": raw = pdu.FirmwareVersion
			case "uptime":           raw = pdu.Uptime
			case "temperature":      raw = pdu.Temperature
			case "humidity":         raw = pdu.Humidity
			default:
				return nil
			}
			v, err := node.NewValue(r.Meta.Type(), raw)
			if err != nil {
				return err
			}
			hnd.Val = v
			return nil
		},
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "input-lines":
				return buildInputLinesContainerNode(pdu.InputLines.Line), nil
			case "outlets":
				return buildOutletsContainerNode(pdu.Outlets.Outlet), nil
			case "sensors":
				return buildSensorsContainerNode(pdu.Sensors.Sensor), nil
			}
			return nil, nil
		},
	}
}

// ---------------------------------------------------------------------------
// input-lines container → line list
// ---------------------------------------------------------------------------

func buildInputLinesContainerNode(lines []InputLine) node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "line" {
				return buildLineListNode(lines), nil
			}
			return nil, nil
		},
	}
}

func buildLineListNode(lines []InputLine) node.Node {
	return &nodeutil.Basic{
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			if r.Row >= len(lines) {
				return nil, nil, nil
			}
			line := lines[r.Row]
			lineNode := &nodeutil.Basic{
				OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
					var raw interface{}
					switch r.Meta.Ident() {
					case "id":        raw = line.ID
					case "voltage":   raw = line.Voltage
					case "current":   raw = line.Current
					case "frequency": raw = line.Frequency
					case "power":     raw = line.Power
					case "energy":    raw = line.Energy
					case "status":    raw = line.Status
					default:          return nil
					}
					v, err := node.NewValue(r.Meta.Type(), raw)
					if err != nil {
						return err
					}
					hnd.Val = v
					return nil
				},
			}
			key, _ := node.NewValues(r.Meta.KeyMeta(), line.ID)
			return lineNode, key, nil
		},
	}
}

// ---------------------------------------------------------------------------
// outlets container → outlet list
// ---------------------------------------------------------------------------

func buildOutletsContainerNode(outlets []Outlet) node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "outlet" {
				return buildOutletListNode(outlets), nil
			}
			return nil, nil
		},
	}
}

func buildOutletListNode(outlets []Outlet) node.Node {
	return &nodeutil.Basic{
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			if r.Row >= len(outlets) {
				return nil, nil, nil
			}
			outlet := outlets[r.Row]
			outletNode := &nodeutil.Basic{
				OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
					var raw interface{}
					switch r.Meta.Ident() {
					case "id":           raw = outlet.ID
					case "label":        raw = outlet.Label
					case "voltage":      raw = outlet.Voltage
					case "current":      raw = outlet.Current
					case "power":        raw = outlet.Power
					case "energy":       raw = outlet.Energy
					case "power-factor": raw = outlet.PowerFactor
					case "status":       raw = outlet.Status
					default:             return nil
					}
					v, err := node.NewValue(r.Meta.Type(), raw)
					if err != nil {
						return err
					}
					hnd.Val = v
					return nil
				},
			}
			key, _ := node.NewValues(r.Meta.KeyMeta(), outlet.ID)
			return outletNode, key, nil
		},
	}
}

// ---------------------------------------------------------------------------
// sensors container → sensor list
// ---------------------------------------------------------------------------

func buildSensorsContainerNode(sensors []Sensor) node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "sensor" {
				return buildSensorListNode(sensors), nil
			}
			return nil, nil
		},
	}
}

func buildSensorListNode(sensors []Sensor) node.Node {
	return &nodeutil.Basic{
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			if r.Row >= len(sensors) {
				return nil, nil, nil
			}
			s := sensors[r.Row]
			sensorNode := &nodeutil.Basic{
				OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
					var raw interface{}
					switch r.Meta.Ident() {
					case "id":     raw = s.ID
					case "type":   raw = s.Type
					case "value":  raw = s.Value
					case "unit":   raw = s.Unit
					case "status": raw = s.Status
					default:       return nil
					}
					v, err := node.NewValue(r.Meta.Type(), raw)
					if err != nil {
						return err
					}
					hnd.Val = v
					return nil
				},
			}
			key, _ := node.NewValues(r.Meta.KeyMeta(), s.ID)
			return sensorNode, key, nil
		},
	}
}
