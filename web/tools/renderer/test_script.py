#!/usr/bin/env python3
import sys
import os
print(f"TEST: Python script started with args: {sys.argv}", file=sys.stderr)
print(f"TEST: Python version: {sys.version}", file=sys.stderr)
print(f"TEST: Current directory: {os.getcwd()}", file=sys.stderr)

try:
    import os
    print(f"TEST: os module imported successfully", file=sys.stderr)
except Exception as e:
    print(f"TEST: Failed to import os: {e}", file=sys.stderr)

try:
    import httpx
    print(f"TEST: httpx imported successfully", file=sys.stderr)
except Exception as e:
    print(f"TEST: Failed to import httpx: {e}", file=sys.stderr)

print("TEST: Script completed successfully")
sys.exit(0)