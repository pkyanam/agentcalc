#!/usr/bin/env python3
"""Replay equivalent named-run and JSON-batch requests against one binary."""
import argparse, importlib.util, json, random, statistics, subprocess, time
from pathlib import Path
HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('evaluate',HERE.parent/'evaluate.py')
evaluate=importlib.util.module_from_spec(spec);spec.loader.exec_module(evaluate)
SHORT='''growth = eval 18750*(1+0.063/12)^37
exact_sum = exact 123456789.123456789 + 0.000000011
combinations = eval choose(52,7)
gib_bytes = convert 3.75 GiB B
fahrenheit_celsius = convert -17.3 f c
trig = eval sin(pi/7)^2 + cos(pi/7)^2 + sqrt(2025)
'''
def main():
    ap=argparse.ArgumentParser();ap.add_argument('--binary',required=True);ap.add_argument('--repeat',type=int,default=100);ap.add_argument('--output');args=ap.parse_args()
    source=(HERE.parent/'runs-v3/arithmetic-cli/solution.sh').read_text()
    batch=source.split("<<'JSONL'\n",1)[1].rsplit('\nJSONL',1)[0]+'\n'
    cases={'run':(['run','--text'],SHORT),'batch':(['batch','--collect','--text'],batch)}
    results={k:{'input_bytes':len(v[1].encode()),'runtime_ms':[]} for k,v in cases.items()}
    rng=random.Random(914);names=list(cases)
    for i in range(args.repeat+3):
        rng.shuffle(names)
        for name in names:
            cmd,data=cases[name];start=time.perf_counter_ns()
            proc=subprocess.run([args.binary,*cmd],input=data,text=True,capture_output=True,timeout=30,check=True)
            elapsed=(time.perf_counter_ns()-start)/1e6
            errors=evaluate.compare(evaluate.gold()['arithmetic'],json.loads(proc.stdout))
            if errors:raise RuntimeError(errors)
            if i>=3:results[name]['runtime_ms'].append(elapsed)
    for r in results.values():r['median_ms']=statistics.median(r['runtime_ms'])
    result={'repetitions':args.repeat,'warmups':3,'results':results,'notes':['Same binary, equivalent arithmetic requests, sequential seeded random order.','Input bytes are payload size, not model-token measurements.']}
    text=json.dumps(result,indent=2)+'\n'
    if args.output:Path(args.output).write_text(text)
    print(text,end='')
if __name__=='__main__':main()
