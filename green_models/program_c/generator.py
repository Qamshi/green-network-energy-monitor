"""
Program C — Agentic Synthetic Data Generator
YANG Model : ietf-isac-utilization (draft-jadoon-green-isac-utilization-03)
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
YANG_FILE    = "./yang/isac-utilization.yang"
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
for a WiFi Access Point performing ISAC (Integrated Sensing and Communication) tasks
such as motion detection in a university hallway.

YANG SCHEMA:
{yang_schema}

Generate a JSON object with a realistic ISAC utilization score. Use these rules:
- overall-utilization-score: 45-75 (active working hours)
- compute-impact-score: usually the highest (50-80)
- energy-consumption-impact-score: correlates with overall score
- memory-impact-score: lower (10-30)
- storage-impact-score: lower (10-25)
- latency-impact-score: moderate (20-45)
- aggregation-window: 60000 (60 seconds in milliseconds)
- score-provenance: "measured"
- sample-count: 100-300
- confidence-level: 80-98
- timestamp: "{timestamp}"

Return ONLY a valid JSON object in this exact structure, no explanation, no markdown:
{{
  "ietf-isac-utilization:isac-utilization": {{
    "overall-utilization-score": <int 0-100>,
    "timestamp": "{timestamp}",
    "aggregation-window": 60000,
    "component-scores": {{
      "compute-impact-score": <int 0-100>,
      "memory-impact-score": <int 0-100>,
      "energy-consumption-impact-score": <int 0-100>,
      "storage-impact-score": <int 0-100>,
      "latency-impact-score": <int 0-100>
    }},
    "metadata": {{
      "score-method": "us-score-method-implementation-specific",
      "score-method-version": "1.0.0",
      "score-provenance": "measured",
      "sample-count": <int>,
      "confidence-level": <int 80-98>
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
    print("  Program C — Agentic Synthetic Data Generator")
    print("  Model: ietf-isac-utilization (IETF GREEN WG)")
    print("  AI Engine: Ollama llama3.1 (local, free)")
    print("=" * 60)

    yang_schema = read_yang_schema()
    raw_text    = generate_data_with_ai(yang_schema)
    data        = validate_and_parse(raw_text)
    save_data(data)

    print()
    print("[Agent] Done! Generated data summary:")
    util = data["ietf-isac-utilization:isac-utilization"]
    print(f"[Agent]   Overall utilization score : {util['overall-utilization-score']}/100")
    scores = util["component-scores"]
    print(f"[Agent]   Compute impact             : {scores['compute-impact-score']}/100")
    print(f"[Agent]   Energy impact              : {scores['energy-consumption-impact-score']}/100")
    print(f"[Agent]   Memory impact              : {scores['memory-impact-score']}/100")
    print(f"[Agent]   Latency impact             : {scores['latency-impact-score']}/100")

if __name__ == "__main__":
    run_agent()
