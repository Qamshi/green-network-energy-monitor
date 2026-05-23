# GREEN Network Energy Monitor
## IETF GREEN Working Group — Multi-Source Energy Data Aggregation

---

## Project Overview

This project implements a network energy monitoring system based on real IETF GREEN
Working Group YANG models. It demonstrates how heterogeneous network devices — each
reporting energy data in different formats — can be standardized and aggregated into
a single unified RESTCONF interface.

### Problem Solved
In real networks, different devices report energy data in incompatible formats.
This project shows how IETF YANG models + RESTCONF can create interoperability
across multiple energy data sources.

---

## Architecture

```
 Ollama AI          Ollama AI          Ollama AI
     |                  |                  |
 Program A          Program B          Program C         Sonoff Device
 Power & Energy     SmartPDU           ISAC Util.        (Real Device)
 port 8881          port 8882          port 8883         port 8080
     |                  |                  |                  |
     └──────────────────┴──────────────────┴──────────────────┘
                                |
                          AGGREGATOR
                          port 8884
                    One unified RESTCONF endpoint
                                |
                       dashboard.html
                      (Web Dashboard)
```

---

## YANG Models Used

| Program | IETF Draft | What it models |
|---------|-----------|----------------|
| Program A | draft-bcmj-green-power-and-energy-yang-04 | Power & energy consumption |
| Program B | draft-ahc-green-smartpdu-yang-00 | Smart Power Distribution Unit |
| Program C | draft-jadoon-green-isac-utilization-03 | ISAC sensing utilization score |
| Sonoff | Custom YANG model (last semester) | Real Sonoff device data |
| Aggregator | Custom aggregator YANG model | Unified view of all sources |

---

## Folder Structure

```
SONOFF/
├── start_all.bat              ← Run this to start everything
├── dashboard.html             ← Open in browser for visual dashboard
├── datastore.json             ← Real Sonoff device data
├── mqtt_collector.py          ← Collects data from Sonoff via MQTT
│
├── restconf/                  ← Last semester Sonoff RESTCONF server
│   └── cmd/rc-test-server/
│       ├── main.go
│       └── sonOff_provider.go
│
└── green_models/              ← This semester's work
    ├── scheduler.py           ← Auto-regenerates data every 60 seconds
    │
    ├── program_a/             ← Power & Energy (port 8881)
    │   ├── yang/power-energy.yang
    │   ├── generator.py       ← Agentic AI data generator
    │   ├── provider.go        ← FreeCONF RESTCONF node
    │   ├── main.go
    │   └── data.json          ← Generated data (auto-updated)
    │
    ├── program_b/             ← SmartPDU (port 8882)
    │   ├── yang/smartpdu.yang
    │   ├── generator.py
    │   ├── provider.go
    │   ├── main.go
    │   └── data.json
    │
    ├── program_c/             ← ISAC Utilization (port 8883)
    │   ├── yang/isac-utilization.yang
    │   ├── generator.py
    │   ├── provider.go
    │   ├── main.go
    │   └── data.json
    │
    └── aggregator/            ← Unified aggregator (port 8884)
        ├── yang/aggregator.yang
        ├── provider.go
        └── main.go
```

---

## How to Run

### Prerequisites
- Go installed
- Python 3 installed
- Ollama installed with llama3.1 model (`ollama pull llama3.1`)

### Start Everything (One Click)
```
Double-click: start_all.bat
```
This opens 6 windows automatically:
- Program A server (port 8881)
- Program B server (port 8882)
- Program C server (port 8883)
- Sonoff server (port 8080)
- Aggregator server (port 8884)
- Scheduler (auto-refreshes data every 60 seconds)

### Open Dashboard
```
Double-click: dashboard.html
```

### RESTCONF Endpoints
| Endpoint | URL |
|----------|-----|
| Program A | http://localhost:8881/restconf/data/power-energy:energy-objects |
| Program B | http://localhost:8882/restconf/data/smartpdu:pdu-system |
| Program C | http://localhost:8883/restconf/data/isac-utilization:isac-utilization |
| Sonoff | http://localhost:8080/restconf/data/sonoff-energy:devices |
| Aggregator | http://localhost:8884/restconf/data/aggregator:aggregated-energy |

---

## How the Agentic AI Works

Each program uses Ollama (llama3.1) running locally as an agentic AI:

1. **Agent reads** the YANG schema file automatically
2. **Agent sends** schema + goal to Llama 3.1
3. **AI autonomously decides** realistic values based on the schema
4. **Agent validates** the response is proper JSON
5. **Agent saves** the result to data.json

This is agentic because the AI reads the schema itself and decides
what realistic data looks like — no values are hardcoded.

---

## Technologies Used

| Technology | Purpose |
|-----------|---------|
| Go + FreeCONF | RESTCONF server (RFC 8040) |
| YANG | Data modelling standard |
| Python + Ollama | Agentic AI data generation |
| Llama 3.1 (local) | Free, local AI model |
| MQTT | Real device data collection (last semester) |
| HTML/CSS/JS | Web dashboard |

---

## Academic Context

- **Last Semester**: Standardized real Sonoff device data using YANG + RESTCONF
- **This Semester**: Implemented 3 IETF GREEN YANG models with agentic AI
  data generation, then aggregated all 4 sources into one unified RESTCONF endpoint

This demonstrates cross-model energy data interoperability — the core goal
of the IETF GREEN Working Group.
