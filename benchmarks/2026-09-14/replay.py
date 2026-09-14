#!/usr/bin/env python3
"""Replay preserved agent-authored recipes in temporary, relocated directories."""
import argparse,json,math,random,shutil,statistics,subprocess,tempfile,time
from pathlib import Path
from evaluate import ROOT,SUITES,ARMS,gold,compare

p=argparse.ArgumentParser()
p.add_argument('--round',choices=['initial','iteration2','final'],default='final')
p.add_argument('--binary',required=True,help='Path to matching agentcalc version')
p.add_argument('--repeat',type=int,default=30)
a=p.parse_args()
if a.repeat<1:p.error('--repeat must be positive')
run_name={'initial':'runs','iteration2':'runs-v2','final':'runs-v3'}[a.round]
expected=gold();names=[f'{s}-{arm}' for s in SUITES for arm in ARMS];times={n:[] for n in names}
with tempfile.TemporaryDirectory(prefix='agentcalc-replay-') as temp:
    root=Path(temp);binary=root/'agentcalc';shutil.copy2(Path(a.binary).expanduser().resolve(),binary);binary.chmod(0o755)
    shutil.copytree(ROOT/'tasks',root/'tasks');shutil.copytree(ROOT/run_name,root/run_name)
    for f in (root/run_name).rglob('*'):
        if f.suffix not in ('.sh','.py','.js','.json') or not f.is_file():continue
        text=f.read_text()
        text=text.replace('/Users/preetham/Code/agentcalc/benchmarks/2026-09-14',str(root))
        for old in ['/Users/preetham/.local/bin/agentcalc','/tmp/agentcalc-benchmark-v2/agentcalc','/tmp/agentcalc-benchmark-v3/agentcalc']:
            text=text.replace(old,str(binary))
        f.write_text(text)
    rng=random.Random(914)
    for i in range(a.repeat+3):
        rng.shuffle(names)
        for name in names:
            start=time.perf_counter_ns()
            process=subprocess.run(['bash',str(root/run_name/name/'solution.sh')],capture_output=True,timeout=30)
            elapsed=(time.perf_counter_ns()-start)/1e6
            if process.returncode:raise RuntimeError(process.stderr.decode())
            errors=compare(expected[name.rsplit('-',1)[0]],json.loads(process.stdout))
            if errors:raise RuntimeError(errors)
            if i>=3:times[name].append(elapsed)
print(json.dumps({n:{'median_ms':statistics.median(v),'p95_ms':sorted(v)[math.ceil(.95*len(v))-1],'correct':True,'repetitions':a.repeat} for n,v in times.items()},indent=2))
