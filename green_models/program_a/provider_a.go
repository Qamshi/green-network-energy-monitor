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
// Data structures matching power-energy.yang
// ---------------------------------------------------------------------------

type PowerData struct {
	InstantaneousPower int    `json:"instantaneous-power"`
	NameplatePower     int    `json:"nameplate-power"`
	UnitMultiplier     string `json:"unit-multiplier"`
	DataSourceAccuracy string `json:"data-source-accuracy"`
	PowerFactor        int    `json:"power-factor"`
	MeasurementLocal   bool   `json:"measurement-local"`
}

type EnergyData struct {
	TotalEnergyConsumed  uint64 `json:"total-energy-consumed"`
	TotalEnergyDelivered uint64 `json:"total-energy-delivered"`
	UnitMultiplier       string `json:"unit-multiplier"`
	DataSourceAccuracy   string `json:"data-source-accuracy"`
	MeasurementLocal     bool   `json:"measurement-local"`
}

type EnergyEntry struct {
	ObjectID string     `json:"object-id"`
	Power    PowerData  `json:"power"`
	Energy   EnergyData `json:"energy"`
}

type DataFile struct {
	EnergyObjects struct {
		EnergyEntry []EnergyEntry `json:"energy-entry"`
	} `json:"ietf-power-and-energy:energy-objects"`
}

// ---------------------------------------------------------------------------
// loadEntries — reads data.json
// ---------------------------------------------------------------------------

func loadEntries() ([]EnergyEntry, error) {
	f, err := os.Open("./data.json")
	if err != nil {
		fmt.Println("ERROR: Could not open data.json")
		return nil, err
	}
	defer f.Close()

	var df DataFile
	if err := json.NewDecoder(f).Decode(&df); err != nil {
		fmt.Printf("ERROR: Failed to decode data.json: %v\n", err)
		return nil, err
	}

	entries := df.EnergyObjects.EnergyEntry
	fmt.Printf("INFO: Loaded %d energy entries from data.json\n", len(entries))
	return entries, nil
}

// ---------------------------------------------------------------------------
// PowerEnergyProvider — root node
//
// YANG tree:
//   module power-energy
//     container energy-objects        <-- OnChild at root level
//       list energy-entry             <-- OnNext inside container node
//         leaf object-id
//         container power
//         container energy
// ---------------------------------------------------------------------------

func PowerEnergyProvider() node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "energy-objects" {
				return energyObjectsNode(), nil
			}
			return nil, nil
		},
	}
}

// energyObjectsNode — represents the container "energy-objects"
// It must handle OnChild for "energy-entry" (the list inside the container)
func energyObjectsNode() node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "energy-entry" {
				return energyEntryListNode(), nil
			}
			return nil, nil
		},
	}
}

// energyEntryListNode — represents the list "energy-entry"
// FreeCONF calls OnNext repeatedly to walk through list rows
func energyEntryListNode() node.Node {
	return &nodeutil.Basic{
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			entries, err := loadEntries()
			if err != nil {
				return nil, nil, err
			}
			if r.Row >= len(entries) {
				return nil, nil, nil
			}
			entry := entries[r.Row]
			key, _ := node.NewValues(r.Meta.KeyMeta(), entry.ObjectID)
			return buildEnergyEntryNode(entry), key, nil
		},
	}
}

func buildEnergyEntryNode(entry EnergyEntry) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			if r.Meta.Ident() == "object-id" {
				v, err := node.NewValue(r.Meta.Type(), entry.ObjectID)
				if err != nil {
					return err
				}
				hnd.Val = v
			}
			return nil
		},
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "power":
				return buildPowerNode(entry.Power), nil
			case "energy":
				return buildEnergyNode(entry.Energy), nil
			}
			return nil, nil
		},
	}
}

func buildPowerNode(p PowerData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "instantaneous-power":
				raw = p.InstantaneousPower
			case "nameplate-power":
				raw = p.NameplatePower
			case "unit-multiplier":
				raw = p.UnitMultiplier
			case "data-source-accuracy":
				raw = p.DataSourceAccuracy
			case "power-factor":
				raw = p.PowerFactor
			case "measurement-local":
				raw = p.MeasurementLocal
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
	}
}

func buildEnergyNode(e EnergyData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "total-energy-consumed":
				raw = e.TotalEnergyConsumed
			case "total-energy-delivered":
				raw = e.TotalEnergyDelivered
			case "unit-multiplier":
				raw = e.UnitMultiplier
			case "data-source-accuracy":
				raw = e.DataSourceAccuracy
			case "measurement-local":
				raw = e.MeasurementLocal
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
	}
}
