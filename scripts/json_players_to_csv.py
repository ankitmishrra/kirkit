#!/usr/bin/env python3
"""Read teams/players JSON and write a CSV with player name and nationality."""
import csv
import json
import sys


def main():
    if len(sys.argv) > 1:
        with open(sys.argv[1], encoding="utf-8") as f:
            data = json.load(f)
    else:
        data = json.load(sys.stdin)

    out_path = sys.argv[2] if len(sys.argv) > 2 else None
    out = open(out_path, "w", newline="", encoding="utf-8") if out_path else sys.stdout
    try:
        writer = csv.writer(out)
        writer.writerow(["player", "nationality"])
        for team in data.get("data", []):
            for p in team.get("players", []):
                writer.writerow([p.get("name", ""), p.get("country", "")])
    finally:
        if out_path:
            out.close()


if __name__ == "__main__":
    main()
