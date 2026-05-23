package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

// ---------------------------------------------------------------------------
// Data structures matching isac-utilization.yang
// ---------------------------------------------------------------------------

type ComponentScores struct {
	ComputeImpact  int `json:"compute-impact-score"`
	MemoryImpact   int `json:"memory-impact-score"`
	EnergyImpact   int `json:"energy-consumption-impact-score"`
	StorageImpact  int `json:"storage-impact-score"`
	LatencyImpact  int `json:"latency-impact-score"`
}

type Metadata struct {
	ScoreMethod        string `json:"score-method"`
	ScoreMethodVersion string `json:"score-method-version"`
	ScoreProvenance    string `json:"score-provenance"`
	SampleCount        uint32 `json:"sample-count"`
	ConfidenceLevel    int    `json:"confidence-level"`
}

type ISACUtil struct {
	OverallScore      int             `json:"overall-utilization-score"`
	Timestamp         string          `json:"timestamp"`
	AggregationWindow uint64          `json:"aggregation-window"`
	ComponentScores   ComponentScores `json:"component-scores"`
	Metadata          Metadata        `json:"metadata"`
}

type ISACDataFile struct {
	ISACUtil ISACUtil `json:"ietf-isac-utilization:isac-utilization"`
}

func loadDatastore() (ISACUtil, error) {
	var empty ISACUtil
	f, err := os.Open("./data.json")
	if err != nil {
		fmt.Println("ERROR: Could not open data.json")
		fmt.Println("Tip: Run 'python generator.py' first.")
		return empty, err
	}
	defer f.Close()

	var df ISACDataFile
	if err := json.NewDecoder(f).Decode(&df); err != nil {
		fmt.Printf("ERROR: Failed to decode data.json: %v\n", err)
		return empty, err
	}
	fmt.Printf("INFO: Loaded ISAC data — overall score: %d/100\n", df.ISACUtil.OverallScore)
	return df.ISACUtil, nil
}

// ---------------------------------------------------------------------------
// ISACProvider — root FreeCONF node
// ---------------------------------------------------------------------------

func ISACProvider() node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "isac-utilization":
				data, err := loadDatastore()
				if err != nil {
					return nil, err
				}
				return buildISACNode(data), nil
			}
			return nil, nil
		},
	}
}

func buildISACNode(d ISACUtil) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "overall-utilization-score": raw = d.OverallScore
			case "timestamp":                 raw = d.Timestamp
			case "aggregation-window":        raw = d.AggregationWindow
			default:                          return nil
			}
			v, err := node.NewValue(r.Meta.Type(), raw)
			if err != nil { return err }
			hnd.Val = v
			return nil
		},
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			switch r.Meta.Ident() {
			case "component-scores":
				return buildComponentScoresNode(d.ComponentScores), nil
			case "metadata":
				return buildMetadataNode(d.Metadata), nil
			}
			return nil, nil
		},
	}
}

func buildComponentScoresNode(cs ComponentScores) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "compute-impact-score":            raw = cs.ComputeImpact
			case "memory-impact-score":             raw = cs.MemoryImpact
			case "energy-consumption-impact-score": raw = cs.EnergyImpact
			case "storage-impact-score":            raw = cs.StorageImpact
			case "latency-impact-score":            raw = cs.LatencyImpact
			default:                                return nil
			}
			v, err := node.NewValue(r.Meta.Type(), raw)
			if err != nil { return err }
			hnd.Val = v
			return nil
		},
	}
}

func buildMetadataNode(m Metadata) node.Node {
	return &nodeutil.Basic{
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			var raw interface{}
			switch r.Meta.Ident() {
			case "score-method":         raw = m.ScoreMethod
			case "score-method-version": raw = m.ScoreMethodVersion
			case "score-provenance":     raw = m.ScoreProvenance
			case "sample-count":         raw = m.SampleCount
			case "confidence-level":     raw = m.ConfidenceLevel
			default:                     return nil
			}
			v, err := node.NewValue(r.Meta.Type(), raw)
			if err != nil { return err }
			hnd.Val = v
			return nil
		},
	}
}
