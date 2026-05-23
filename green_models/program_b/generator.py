"""
Program B — Agentic Synthetic Data Generator
YANG Model : pdu-common (draft-ahc-green-smartpdu-yang-00)
AI Engine  : Ollama (llama3.1) — runs locally, completely free

How it works (Agentic behaviour):
  Step 1 — Agent reads the YANG schema file automatically
  Step 2 — Agent sends schema + goal to Llama 3.1
  Step 3 — AI autonomously decides realistic values for each field
  Step 4 — Agent validates the response is proper JSON
  Step 5 — Agent saves the result to data.json
"""

import json
import requests
import os
from datetime import datetime, timezone

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
OLLAMA_URL   = "http://localhost:11434/api/generate"
OLLAMA_MODEL = "llama3.1"
YANG_FILE    = "./yang/smartpdu.yang"
OUTPUT_FILE  = "./data.json"

# ---------------------------------------------------------------------------
# Step 1 — Agent reads the YANG schema automatically
# ---------------------------------------------------------------------------
def read_yang_schema():
    print("[Agent] Step 1: Reading YANG schema from", YANG_FILE)
    if not os.path.exists(YANG_FILE):
        print("[Agent] ERROR: YANG file not found at", YANG_FILE)
        exit(1)
    with open(YANG_FILE, "r") as f:
        schema = f.read()
    print("[Agent] YANG schema loaded successfully.")
    return schema

# ---------------------------------------------------------------------------
# Step 2 & 3 — Agent sends schema to Ollama, AI generates realistic data
# ---------------------------------------------------------------------------
def generate_data_with_ai(yang_schema):
    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    print(f"[Agent] Step 2: Sending YANG schema to {OLLAMA_MODEL}...")
    print(f"[Agent] Step 3: AI is generating realistic synthetic data...")

    prompt = f"""You are a network device data simulator.

I will give you a YANG model schema. Your job is to generate realistic synthetic data
for a Smart Power Distribution Unit (SmartPDU) in a university network room.
The PDU powers a WiFi Access Point, a Router, and a Network Switch.

YANG SCHEMA:
{yang_schema}

Generate a JSON object with realistic SmartPDU readings. Use these rules:
- manufacturer: "APC", model: "AP8941"
- 1 input line (L1) with European voltage (230V, 50Hz)
- 3 outlets: outlet-1 (Access-Point ~12W), outlet-2 (Router ~20W), outlet-3 (Switch ~18W)
- 2 sensors: temperature (~24 Celsius) and humidity (~42%)
- uptime: 2-5 days in seconds (172800 to 432000)
- power-factor: decimal between 0.92 and 0.99
- all outlets status "on"

Return ONLY a valid JSON object in this exact structure, no explanation, no markdown:
{{
  "pdu-common:pdu-system": {{
    "manufacturer": "APC",
    "model": "AP8941",
    "serial-number": "<string>",
    "firmware-version": "<string>",
    "uptime": <int>,
    "temperature": <decimal>,
    "humidity": <decimal>,
    "input-lines": {{
      "line": [
        {{
          "id": "L1",
          "voltage": <decimal>,
          "current": <decimal>,
          "frequency": 50.0,
          "power": <decimal>,
          "energy": <decimal>,
          "status": "ok"
        }}
      ]
    }},
    "outlets": {{
      "outlet": [
        {{
          "id": "outlet-1",
          "label": "Access-Point",
          "voltage": <decimal>,
          "current": <decimal>,
          "power": <decimal>,
          "energy": <decimal>,
          "power-factor": <decimal>,
          "status": "on"
        }},
        {{
          "id": "outlet-2",
          "label": "Router",
          "voltage": <decimal>,
          "current": <decimal>,
          "power": <decimal>,
          "energy": <decimal>,
          "power-factor": <decimal>,
          "status": "on"
        }},
        {{
          "id": "outlet-3",
          "label": "Switch",
          "voltage": <decimal>,
          "current": <decimal>,
          "power": <decimal>,
          "energy": <decimal>,
          "power-factor": <decimal>,
          "status": "on"
        }}
      ]
    }},
    "sensors": {{
      "sensor": [
        {{
          "id": "sensor-temp-1",
          "type": "temperature",
          "value": <decimal>,
          "unit": "Celsius",
          "status": "ok"
        }},
        {{
          "id": "sensor-humidity-1",
          "type": "humidity",
          "value": <decimal>,
          "unit": "%",
          "status": "ok"
        }}
      ]
    }}
  }}
}}"""

    response = requests.post(OLLAMA_URL, json={
        "model": OLLAMA_MODEL,
        "prompt": prompt,
        "stream": False
    })

    if response.status_code != 200:
        print("[Agent] ERROR: Ollama returned status", response.status_code)
        exit(1)

    raw_text = response.json()["response"].strip()
    print("[Agent] AI response received.")
    return raw_text

# ---------------------------------------------------------------------------
# Step 4 — Agent validates the JSON response
# ---------------------------------------------------------------------------
def validate_and_parse(raw_text):
    print("[Agent] Step 4: Validating AI response...")

    # Strip markdown code blocks if AI added them
    if "```json" in raw_text:
        raw_text = raw_text.split("```json")[1].split("```")[0].strip()
    elif "```" in raw_text:
        raw_text = raw_text.split("```")[1].split("```")[0].strip()

    # Find the JSON object in the response
    start = raw_text.find("{")
    end   = raw_text.rfind("}") + 1
    if start == -1 or end == 0:
        print("[Agent] ERROR: No JSON found in AI response.")
        print("[Agent] Raw response was:", raw_text)
        exit(1)

    json_text = raw_text[start:end]

    try:
        data = json.loads(json_text)
        print("[Agent] JSON validation passed.")
        return data
    except json.JSONDecodeError as e:
        print("[Agent] ERROR: Invalid JSON from AI:", e)
        print("[Agent] Raw JSON was:", json_text)
        exit(1)

# ---------------------------------------------------------------------------
# Step 5 — Agent saves to data.json
# ---------------------------------------------------------------------------
def save_data(data):
    print("[Agent] Step 5: Saving data to", OUTPUT_FILE)
    with open(OUTPUT_FILE, "w") as f:
        json.dump(data, f, indent=2)
    print("[Agent] data.json saved successfully!")

# ---------------------------------------------------------------------------
# Main — run the agent
# ---------------------------------------------------------------------------
def run_agent():
    print("=" * 60)
    print("  Program B — Agentic Synthetic Data Generator")
    print("  Model: pdu-common (IETF GREEN WG)")
    print("  AI Engine: Ollama llama3.1 (local, free)")
    print("=" * 60)

    yang_schema = read_yang_schema()
    raw_text    = generate_data_with_ai(yang_schema)
    data        = validate_and_parse(raw_text)
    save_data(data)

    print()
    print("[Agent] Done! Generated data summary:")
    pdu = data["pdu-common:pdu-system"]
    outlets = pdu["outlets"]["outlet"]
    for o in outlets:
        print(f"[Agent]   {o['label']}: {o['power']}W")
    total = sum(o["power"] for o in outlets)
    print(f"[Agent]   Total PDU load: {total:.1f}W")

if __name__ == "__main__":
    run_agent()
