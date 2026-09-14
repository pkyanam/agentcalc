#!/usr/bin/env python3
"""Print fresh-agent launch specifications; does not invoke an agent service."""
import argparse, json
from pathlib import Path
ap=argparse.ArgumentParser();ap.add_argument('--round',type=int,required=True);ap.add_argument('--binary',required=True);ap.add_argument('--root',type=Path,default=Path(__file__).resolve().parents[3]);args=ap.parse_args()
root=args.root.resolve();base=root/'benchmarks/2026-09-14'
items=[]
for suite in ['arithmetic','data','numerical']:
 for arm in ['ordinary','cli']:
  start='Controlled CLI benchmark.' if arm=='cli' else 'Controlled ordinary-tools benchmark. Do not read ANY skill file (including installed agentcalc) or invoke agentcalc.'
  message=f'{start} Read together ONLY {base}/optimized-routing/protocol.txt and {base}/tasks/{suite}.txt'
  if arm=='cli':message+=f' plus {root}/skills/agentcalc/SKILL.md'
  message+=', then solve.'
  if suite=='data':message+=f' CSV: {base}/tasks/sales.csv.'
  if arm=='cli':message+=f' Binary: {args.binary}.'
  message+=' Work efficiently and stop after successful calculation. Final exactly requested JSON; it is delivered automatically, so do not discover or call communication tools or make unrelated calls. No artifacts.'
  items.append({'task_name':f'route{args.round}_{suite}_{arm}','model':'gpt-5.6-luna','reasoning_effort':'medium','fork_turns':'none','message':message})
print(json.dumps(items,indent=2))
