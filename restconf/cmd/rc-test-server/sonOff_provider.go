package main

import (
	"encoding/json"
	"fmt" // Add this for logging
	"os"

	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/val"
)

func loadDatastore() (map[string]interface{}, error) {
	// Use the absolute path to be 100% sure it finds the file
	path := "C:/QAMROSH MAQSOOD/Local Disk E/Master in IT (IOT)/sonoff/datastore.json"
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("ERROR: Could not find datastore.json at %s\n", path)
		return nil, err
	}
	defer f.Close()

	var raw map[string]interface{}
	err = json.NewDecoder(f).Decode(&raw)
	if err != nil {
		fmt.Printf("ERROR: Failed to decode JSON: %s\n", err)
	}
	return raw, err
}

func DevicesProvider() node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "device" {
				return DeviceListProvider(), nil
			}
			return nil, nil
		},
	}
}

func DeviceListProvider() node.Node {
	return &nodeutil.Basic{
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			raw, err := loadDatastore()
			if err != nil {
				return nil, nil, err
			}

			var items []map[string]interface{}
			for devID, devData := range raw {
				items = append(items, map[string]interface{}{
					"device-id": devID,
					"sensor":    devData,
				})
			}

			if r.Row >= len(items) {
				return nil, nil, nil
			}

			item := items[r.Row]

			deviceNode := &nodeutil.Basic{
				OnField: func(field node.FieldRequest, hnd *node.ValueHandle) error {
					if field.Meta.Ident() == "device-id" {
						var err error
						hnd.Val, err = node.NewValue(field.Meta.Type(), item["device-id"])
						return err
					}
					return nil
				},
				OnChild: func(child node.ChildRequest) (node.Node, error) {
					if child.Meta.Ident() == "sensor" {
						if sensorData, ok := item["sensor"].(map[string]interface{}); ok {
							// Use ReflectChild instead of Reflect to fix the conversion error
							return nodeutil.ReflectChild(sensorData), nil
						}
					}
					return nil, nil
				},
			}

			key, _ := node.NewValues(r.Meta.KeyMeta(), item["device-id"])
			return deviceNode, key, nil
		},
	}
}