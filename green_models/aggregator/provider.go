package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

// ---------------------------------------------------------------------------
// Source URLs
// ---------------------------------------------------------------------------
const (
	urlProgramA = "http://localhost:8881/restconf/data/power-energy:energy-objects"
	urlProgramB = "http://localhost:8882/restconf/data/smartpdu:pdu-system"
	urlProgramC = "http://localhost:8883/restconf/data/isac-utilization:isac-utilization"
	urlSonoff   = "http://localhost:8080/restconf/data/sonoff-energy:devices"
)

// ---------------------------------------------------------------------------
// Aggregated data structs
// ---------------------------------------------------------------------------

type PowerEnergyData struct {
	ObjectID           string `json:"object-id"`
	InstantaneousPower uint64 `json:"instantaneous-power"`
	NameplatePower     uint64 `json:"nameplate-power"`
	TotalEnergyConsumed uint64 `json:"total-energy-consumed"`
	PowerFactor        uint8  `json:"power-factor"`
	UnitMultiplier     string `json:"unit-multiplier"`
}

type SmartPDUData struct {
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	Temperature   float64 `json:"temperature"`
	Humidity      float64 `json:"humidity"`
	TotalPDUPower float64 `json:"total-pdu-power"`
	ActiveOutlets uint8   `json:"active-outlets"`
}

type ISACData struct {
	OverallScore      uint8  `json:"overall-utilization-score"`
	ComputeImpact     uint8  `json:"compute-impact-score"`
	EnergyImpact      uint8  `json:"energy-impact-score"`
	ConfidenceLevel   uint8  `json:"confidence-level"`
	AggregationWindow uint64 `json:"aggregation-window"`
}

type SonoffData struct {
	DeviceID    string  `json:"device-id"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	PowerActive float64 `json:"power-active"`
	PowerFactor float64 `json:"power-factor"`
	EnergyTotal float64 `json:"energy-total"`
	LastUpdated string  `json:"last-updated"`
}

type AggregatedData struct {
	Timestamp    string
	TotalSources uint8
	PowerEnergy  PowerEnergyData
	SmartPDU     SmartPDUData
	ISAC         ISACData
	Sonoff       SonoffData
}

// ---------------------------------------------------------------------------
// HTTP helper — fetch JSON from a URL
// ---------------------------------------------------------------------------
func fetchJSON(url string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON from %s: %v", url, err)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Fetch & parse each source
// ---------------------------------------------------------------------------

func fetchProgramA() (PowerEnergyData, error) {
	var d PowerEnergyData
	raw, err := fetchJSON(urlProgramA)
	if err != nil {
		return d, err
	}

	entries, ok := raw["energy-entry"].([]interface{})
	if !ok || len(entries) == 0 {
		return d, fmt.Errorf("program A: no energy-entry found")
	}

	entry := entries[0].(map[string]interface{})
	d.ObjectID, _ = entry["object-id"].(string)

	if power, ok := entry["power"].(map[string]interface{}); ok {
		if v, ok := power["instantaneous-power"].(float64); ok {
			d.InstantaneousPower = uint64(v)
		}
		if v, ok := power["nameplate-power"].(float64); ok {
			d.NameplatePower = uint64(v)
		}
		if v, ok := power["power-factor"].(float64); ok {
			d.PowerFactor = uint8(v)
		}
		d.UnitMultiplier, _ = power["unit-multiplier"].(string)
	}
	if energy, ok := entry["energy"].(map[string]interface{}); ok {
		if v, ok := energy["total-energy-consumed"].(float64); ok {
			d.TotalEnergyConsumed = uint64(v)
		}
	}

	fmt.Printf("[Aggregator] Program A fetched — device: %s, power: %d mW\n",
		d.ObjectID, d.InstantaneousPower)
	return d, nil
}

func fetchProgramB() (SmartPDUData, error) {
	var d SmartPDUData
	raw, err := fetchJSON(urlProgramB)
	if err != nil {
		return d, err
	}

	d.Manufacturer, _ = raw["manufacturer"].(string)
	d.Model, _ = raw["model"].(string)
	d.Temperature, _ = raw["temperature"].(float64)
	d.Humidity, _ = raw["humidity"].(float64)

	if outlets, ok := raw["outlets"].(map[string]interface{}); ok {
		if outletList, ok := outlets["outlet"].([]interface{}); ok {
			var totalPower float64
			for _, o := range outletList {
				outlet := o.(map[string]interface{})
				if status, _ := outlet["status"].(string); status == "on" {
					d.ActiveOutlets++
					if p, ok := outlet["power"].(float64); ok {
						totalPower += p
					}
				}
			}
			d.TotalPDUPower = totalPower
		}
	}

	fmt.Printf("[Aggregator] Program B fetched — PDU: %s %s, total power: %.1fW\n",
		d.Manufacturer, d.Model, d.TotalPDUPower)
	return d, nil
}

func fetchProgramC() (ISACData, error) {
	var d ISACData
	raw, err := fetchJSON(urlProgramC)
	if err != nil {
		return d, err
	}

	if v, ok := raw["overall-utilization-score"].(float64); ok {
		d.OverallScore = uint8(v)
	}
	if v, ok := raw["aggregation-window"].(float64); ok {
		d.AggregationWindow = uint64(v)
	}
	if scores, ok := raw["component-scores"].(map[string]interface{}); ok {
		if v, ok := scores["compute-impact-score"].(float64); ok {
			d.ComputeImpact = uint8(v)
		}
		if v, ok := scores["energy-consumption-impact-score"].(float64); ok {
			d.EnergyImpact = uint8(v)
		}
	}
	if meta, ok := raw["metadata"].(map[string]interface{}); ok {
		if v, ok := meta["confidence-level"].(float64); ok {
			d.ConfidenceLevel = uint8(v)
		}
	}

	fmt.Printf("[Aggregator] Program C fetched — utilization score: %d/100\n",
		d.OverallScore)
	return d, nil
}

func fetchSonoff() (SonoffData, error) {
	var d SonoffData
	raw, err := fetchJSON(urlSonoff)
	if err != nil {
		return d, err
	}

	// Sonoff returns: {"device": [{"device-id": "sonoff-001", "sensor": {...}}]}
	deviceList, ok := raw["device"].([]interface{})
	if !ok || len(deviceList) == 0 {
		return d, fmt.Errorf("sonoff: no device found")
	}

	dev := deviceList[0].(map[string]interface{})
	d.DeviceID, _ = dev["device-id"].(string)

	if sensor, ok := dev["sensor"].(map[string]interface{}); ok {
		// Each field is {"value": X, "unit": Y}
		if v, ok := sensor["voltage"].(map[string]interface{}); ok {
			d.Voltage, _ = v["value"].(float64)
		}
		if v, ok := sensor["current"].(map[string]interface{}); ok {
			d.Current, _ = v["value"].(float64)
		}
		// Note: Sonoff uses underscore: power_active, power_factor, energy_total
		if v, ok := sensor["power_active"].(map[string]interface{}); ok {
			d.PowerActive, _ = v["value"].(float64)
		}
		if v, ok := sensor["power_factor"].(map[string]interface{}); ok {
			d.PowerFactor, _ = v["value"].(float64)
		}
		if v, ok := sensor["energy_total"].(map[string]interface{}); ok {
			d.EnergyTotal, _ = v["value"].(float64)
		}
		if v, ok := sensor["timestamp"].(map[string]interface{}); ok {
			d.LastUpdated, _ = v["value"].(string)
		}
	}

	fmt.Printf("[Aggregator] Sonoff fetched — device: %s, power: %.1fW\n",
		d.DeviceID, d.PowerActive)
	return d, nil
}

// ---------------------------------------------------------------------------
// Aggregate all sources
// ---------------------------------------------------------------------------

func aggregate() AggregatedData {
	fmt.Println("[Aggregator] Fetching data from all 4 sources...")
	agg := AggregatedData{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		TotalSources: 4,
	}

	if d, err := fetchProgramA(); err != nil {
		fmt.Println("[Aggregator] WARNING: Program A unavailable:", err)
	} else {
		agg.PowerEnergy = d
	}

	if d, err := fetchProgramB(); err != nil {
		fmt.Println("[Aggregator] WARNING: Program B unavailable:", err)
	} else {
		agg.SmartPDU = d
	}

	if d, err := fetchProgramC(); err != nil {
		fmt.Println("[Aggregator] WARNING: Program C unavailable:", err)
	} else {
		agg.ISAC = d
	}

	if d, err := fetchSonoff(); err != nil {
		fmt.Println("[Aggregator] WARNING: Sonoff unavailable:", err)
	} else {
		agg.Sonoff = d
	}

	fmt.Println("[Aggregator] Aggregation complete.")
	return agg
}

// ---------------------------------------------------------------------------
// AggregatorProvider — FreeCONF node tree
// ---------------------------------------------------------------------------

func AggregatorProvider() node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			if r.Meta.Ident() == "aggregated-energy" {
				data := aggregate()
				return buildAggregatedNode(data), nil
			}
			return nil, nil
		},
	}
}

func buildAggregatedNode(d AggregatedData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "timestamp":     raw = d.Timestamp
			case "total-sources": raw = d.TotalSources
			default:              return nil
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
			case "power-energy":
				return buildPowerEnergyNode(d.PowerEnergy), nil
			case "smart-pdu":
				return buildSmartPDUNode(d.SmartPDU), nil
			case "isac-utilization":
				return buildISACNode(d.ISAC), nil
			case "sonoff-device":
				return buildSonoffNode(d.Sonoff), nil
			}
			return nil, nil
		},
	}
}

func buildPowerEnergyNode(d PowerEnergyData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "object-id":             raw = d.ObjectID
			case "instantaneous-power":   raw = d.InstantaneousPower
			case "nameplate-power":       raw = d.NameplatePower
			case "total-energy-consumed": raw = d.TotalEnergyConsumed
			case "power-factor":          raw = d.PowerFactor
			case "unit-multiplier":       raw = d.UnitMultiplier
			default:                      return nil
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

func buildSmartPDUNode(d SmartPDUData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "manufacturer":   raw = d.Manufacturer
			case "model":          raw = d.Model
			case "temperature":    raw = d.Temperature
			case "humidity":       raw = d.Humidity
			case "total-pdu-power": raw = d.TotalPDUPower
			case "active-outlets": raw = d.ActiveOutlets
			default:               return nil
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

func buildISACNode(d ISACData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "overall-utilization-score": raw = d.OverallScore
			case "compute-impact-score":      raw = d.ComputeImpact
			case "energy-impact-score":       raw = d.EnergyImpact
			case "confidence-level":          raw = d.ConfidenceLevel
			case "aggregation-window":        raw = d.AggregationWindow
			default:                          return nil
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

func buildSonoffNode(d SonoffData) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "device-id":    raw = d.DeviceID
			case "voltage":      raw = d.Voltage
			case "current":      raw = d.Current
			case "power-active": raw = d.PowerActive
			case "power-factor": raw = d.PowerFactor
			case "energy-total": raw = d.EnergyTotal
			case "last-updated": raw = d.LastUpdated
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
}
