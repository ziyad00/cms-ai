from __future__ import annotations

import json
import sys
from pathlib import Path

from cms_ai.spec_validator import validate_template_spec


def main(argv: list[str] | None = None) -> int:
    argv = sys.argv[1:] if argv is None else argv
    if len(argv) != 1:
        print("Usage: python -m cms_ai.validate <spec.json>", file=sys.stderr)
        return 2

    spec_path = Path(argv[0])
    try:
        data = json.loads(spec_path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        print(f"File not found: {spec_path}", file=sys.stderr)
        return 2
    except json.JSONDecodeError as e:
        print(f"Invalid JSON: {e}", file=sys.stderr)
        return 2

    errors = validate_template_spec(data)
    if errors:
        for error in errors:
            print(f"{error.path}: {error.message}", file=sys.stderr)
        return 1

    print("OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
