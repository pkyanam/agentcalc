#!/usr/bin/env python3
"""Collect sanitized usage and validate answer-only finals for routing runs.

Session logs are read locally. Message text is used transiently to parse the
last assistant JSON object and is never written to the output.
"""
from __future__ import annotations

import argparse
import importlib.util
import json
import re
from pathlib import Path

HERE = Path(__file__).resolve().parent
EVAL_PATH = HERE.parent / "evaluate.py"
spec = importlib.util.spec_from_file_location("benchmark_evaluate", EVAL_PATH)
evaluate = importlib.util.module_from_spec(spec)
assert spec.loader is not None
spec.loader.exec_module(evaluate)


def _text_parts(value):
    if isinstance(value, str):
        yield value
    elif isinstance(value, list):
        for x in value:
            yield from _text_parts(x)
    elif isinstance(value, dict):
        # These are response message fields; deliberately ignore tool inputs.
        for k in ("text", "output_text", "content"):
            if k in value:
                yield from _text_parts(value[k])


def final_json(events):
    """Return (object, format_error) for the last assistant response item."""
    for event in reversed(events):
        if event.get("type") != "response_item":
            continue
        p = event.get("payload", {})
        if p.get("type") not in ("message", "output_message"):
            continue
        if p.get("role") != "assistant" or (p.get("channel") != "final" and p.get("phase") != "final_answer"):
            continue
        raw = "".join(_text_parts(p.get("content", p.get("output", p)))) .strip()
        if not raw:
            continue
        try:
            value = json.loads(raw)
        except (ValueError, TypeError):
            return None, "final response was not exactly one JSON value"
        if not isinstance(value, dict):
            return None, "final JSON value was not an object"
        return value, None
    return None, "no assistant response_item final found"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--sessions", required=True)
    ap.add_argument("--prefix", default="route")
    ap.add_argument("--rounds", default="5,6", help="comma-separated route numbers, or all")
    ap.add_argument("--output")
    args = ap.parse_args()

    selected = None if args.rounds == "all" else {int(x) for x in args.rounds.split(",") if x}
    # Reuse the established usage collector; it deduplicates response IDs and
    # returns only counters, timestamps, model metadata, and tool-call metadata.
    base = evaluate.collect(args.sessions, args.prefix)
    expected = evaluate.gold()
    records = {}
    for path in sorted(Path(args.sessions).glob("*.jsonl")):
        with path.open() as stream:
            try:
                first = json.loads(next(stream))
            except (StopIteration, ValueError):
                continue
            agent_path = first.get("payload", {}).get("agent_path", "")
            if not agent_path.startswith("/root/" + args.prefix):
                continue
            events = []
            for line in stream:
                try:
                    events.append(json.loads(line))
                except ValueError:
                    continue
            name = agent_path.rsplit("/", 1)[-1]
            m = re.match(r"route(\d+)_", name)
            round_number = int(m.group(1)) if m else None
            if selected is not None and round_number not in selected:
                continue
            # evaluate.collect's key normalizes the same path for a lookup.
            base_key = agent_path.removeprefix("/root/" + args.prefix).replace("_", "-")
            metric = base.get(base_key, {})
            final, format_error = final_json(events)
            suite = next((s for s in ("arithmetic", "data", "numerical", "tiny") if s in name), None)
            errors = []
            if format_error:
                errors.append(format_error)
            elif suite in expected:
                errors.extend(evaluate.compare(expected[suite], final))
            elif suite == "tiny":
                if final != {"answer": 8}:
                    errors.append("tiny answer mismatch")
            else:
                errors.append("could not identify suite")
            completion_errors = []
            if not metric.get("tokens"):
                completion_errors.append("missing token usage metric")
            if not any(e.get("type") == "event_msg" and e.get("payload", {}).get("type") == "task_complete" for e in events):
                completion_errors.append("task did not complete")
            if format_error:
                completion_errors.append("missing valid final channel JSON")
            # Keep only sanitized answer facts, never the answer or transcript.
            records[name] = {
                "suite": suite,
                "arm": "cli" if "cli" in name.split("_") else "ordinary",
                "correct": not errors,
                "correctness_errors": errors,
                "completed": not completion_errors,
                "completion_errors": completion_errors,
                "round": round_number,
                "model": metric.get("model"),
                "effort": metric.get("effort"),
                "model_responses": metric.get("model_responses"),
                "outer_tool_calls": metric.get("outer_tool_calls"),
                "wall_seconds": metric.get("wall_seconds"),
                "tokens": metric.get("tokens", {}),
                "usage_records": metric.get("usage_records", []),
            }
    payload = {
        "protocol": "optimized-routing",
        "prefix": args.prefix,
        "rounds": "all" if selected is None else sorted(selected),
        "agents": records,
        "notes": [
            "Usage is per-response and deduplicated by response ID via evaluate.collect.",
            "uncached_input_tokens = input_tokens - cached_input_tokens.",
            "Final answer text and hidden reasoning are not stored.",
        ],
    }
    rendered = json.dumps(payload, indent=2, sort_keys=True) + "\n"
    if args.output:
        Path(args.output).write_text(rendered)
    print(rendered, end="")


if __name__ == "__main__":
    main()
