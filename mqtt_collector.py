import json
from pathlib import Path
from datetime import datetime, timezone
import paho.mqtt.client as mqtt

# ==============================
# Configuration
# ==============================
DATASTORE_FILE = Path("datastore.json")

BROKER = "172.20.10.2"
PORT = 1883
TOPIC = "tele/+/SENSOR"

# Alert thresholds
THRESHOLDS = {
    "voltage_max": 250,
    "current_max": 1.0,
    "power_max": 100
}

# ==============================
# MQTT callbacks
# ==============================
def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("✅ Connected to MQTT broker")
        client.subscribe(TOPIC)
    else:
        print("❌ MQTT connection failed:", rc)

def on_message(client, userdata, msg):
    try:
        payload = json.loads(msg.payload.decode())
    except json.JSONDecodeError:
        print("❌ Invalid JSON:", msg.payload.decode())
        return

    energy = payload.get("ENERGY")
    if not energy:
        return

    device_id = msg.topic.split("/")[1]
    now = datetime.now(timezone.utc).isoformat()

    # Load datastore
    if DATASTORE_FILE.exists():
        with open(DATASTORE_FILE, "r") as f:
            devices = json.load(f)
    else:
        devices = {}

    # Ensure device structure exists
    devices.setdefault(device_id, {})
    device = devices[device_id]

    # Initialize alerts
    device.setdefault("alerts", [])

    # ==============================
    # Latest standardized data
    # ==============================
    device.update({
        "timestamp": {"value": payload.get("Time", now), "unit": "UTC"},
        "voltage": {"value": energy.get("Voltage", 0), "unit": "V"},
        "current": {"value": energy.get("Current", 0), "unit": "A"},
        "power_active": {"value": energy.get("Power", 0), "unit": "W"},
        "power_apparent": {"value": energy.get("ApparentPower", 0), "unit": "VA"},
        "power_reactive": {"value": energy.get("ReactivePower", 0), "unit": "var"},
        "power_factor": {"value": energy.get("Factor", 0), "unit": "ratio"},
        "energy_today": {"value": energy.get("Today", 0), "unit": "kWh"},
        "energy_yesterday": {"value": energy.get("Yesterday", 0), "unit": "kWh"},
        "energy_total": {"value": energy.get("Total", 0), "unit": "kWh"}
    })

    # ==============================
    # Alerts
    # ==============================
    device["alerts"] = []  # reset alerts each update
    if energy.get("Voltage", 0) > THRESHOLDS["voltage_max"]:
        device["alerts"].append({
            "time": now,
            "type": "OVER_VOLTAGE",
            "message": f"Voltage exceeded {THRESHOLDS['voltage_max']} V"
        })
    if energy.get("Current", 0) > THRESHOLDS["current_max"]:
        device["alerts"].append({
            "time": now,
            "type": "OVER_CURRENT",
            "message": f"Current exceeded {THRESHOLDS['current_max']} A"
        })
    if energy.get("Power", 0) > THRESHOLDS["power_max"]:
        device["alerts"].append({
            "time": now,
            "type": "OVER_POWER",
            "message": f"Power exceeded {THRESHOLDS['power_max']} W"
        })

    # Save updated datastore
    with open(DATASTORE_FILE, "w") as f:
        json.dump(devices, f, indent=2)

    print(f"📥 Updated {device_id}")

# ==============================
# MQTT Client
# ==============================
client = mqtt.Client(client_id="SonoffCollector")
client.on_connect = on_connect
client.on_message = on_message

client.connect(BROKER, PORT, 60)
client.loop_forever()
