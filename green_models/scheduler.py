"""
AUTO-SCHEDULER — Runs all 3 generators automatically every 60 seconds
Place this file inside: green_models/
Run: python scheduler.py

What it does:
  Every 60 seconds:
    - Calls Program A generator → updates program_a/data.json
    - Calls Program B generator → updates program_b/data.json
    - Calls Program C generator → updates program_c/data.json

The RESTCONF servers (go run .) read data.json on every request,
so the dashboard automatically shows fresh data on each refresh.
"""

import time
import subprocess
import sys
import os
from datetime import datetime, timezone

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
INTERVAL_SECONDS = 60  # How often to regenerate data

# Paths to each generator (relative to this file's location)
BASE_DIR = os.path.dirname(os.path.abspath(__file__))

GENERATORS = [
    {
        "name": "Program A — Power & Energy",
        "path": os.path.join(BASE_DIR, "program_a", "generator.py"),
        "cwd":  os.path.join(BASE_DIR, "program_a"),
    },
    {
        "name": "Program B — SmartPDU",
        "path": os.path.join(BASE_DIR, "program_b", "generator.py"),
        "cwd":  os.path.join(BASE_DIR, "program_b"),
    },
    {
        "name": "Program C — ISAC Utilization",
        "path": os.path.join(BASE_DIR, "program_c", "generator.py"),
        "cwd":  os.path.join(BASE_DIR, "program_c"),
    },
]

# ---------------------------------------------------------------------------
# Run a single generator
# ---------------------------------------------------------------------------
def run_generator(gen):
    try:
        result = subprocess.run(
            [sys.executable, gen["path"]],
            cwd=gen["cwd"],
            capture_output=True,
            text=True,
            timeout=120  # 2 minute timeout per generator
        )
        if result.returncode == 0:
            print(f"    ✓ {gen['name']} — data updated")
        else:
            print(f"    ✗ {gen['name']} — ERROR:")
            print(f"      {result.stderr.strip()}")
    except subprocess.TimeoutExpired:
        print(f"    ✗ {gen['name']} — TIMEOUT (Ollama may be slow)")
    except Exception as e:
        print(f"    ✗ {gen['name']} — EXCEPTION: {e}")

# ---------------------------------------------------------------------------
# Run all generators once
# ---------------------------------------------------------------------------
def run_all():
    now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    print(f"\n[Scheduler] Generating fresh data at {now}")
    print("-" * 50)
    for gen in GENERATORS:
        run_generator(gen)
    print("-" * 50)
    print(f"[Scheduler] Done. Next update in {INTERVAL_SECONDS} seconds...")

# ---------------------------------------------------------------------------
# Main loop
# ---------------------------------------------------------------------------
def main():
    print("=" * 60)
    print("  GREEN Energy Data Scheduler")
    print(f"  Refreshing all 3 generators every {INTERVAL_SECONDS} seconds")
    print("  Press Ctrl+C to stop")
    print("=" * 60)

    # Run immediately on start
    run_all()

    # Then keep running on interval
    while True:
        try:
            time.sleep(INTERVAL_SECONDS)
            run_all()
        except KeyboardInterrupt:
            print("\n[Scheduler] Stopped by user.")
            break

if __name__ == "__main__":
    main()
