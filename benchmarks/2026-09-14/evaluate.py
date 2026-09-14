#!/usr/bin/env python3
"""Independent correctness, sanitized usage, and process replay measurements.

No transcript text is copied. Pass --sessions to collect metadata for agent paths
/root/bench_<suite>_<arm>. Use --prefix for later fresh-agent rounds.
"""
import argparse, csv, hashlib, itertools, json, math, platform, random, statistics
import subprocess, time
from datetime import datetime
from decimal import Decimal
from fractions import Fraction
from pathlib import Path

ROOT = Path(__file__).resolve().parent
SUITES = ('arithmetic','data','numerical')
ARMS = ('cli','ordinary')

def gold():
    a = {
      'growth':18750*(1+.063/12)**37,
      'exact_sum':str(Fraction('123456789.123456789')+Fraction('0.000000011')),
      'combinations':math.comb(52,7), 'gib_bytes':3.75*1024**3,
      'fahrenheit_celsius':(-17.3-32)*5/9,
      'trig':math.sin(math.pi/7)**2+math.cos(math.pi/7)**2+math.sqrt(2025)}
    with (ROOT/'tasks/sales.csv').open() as f:
        rows=list(csv.DictReader(f))
    r=sorted(Decimal(x['revenue']) for x in rows)
    def percentile(p):
        pos=Decimal(len(r)-1)*Decimal(p); i=int(pos)
        return float(r[i]+(r[min(i+1,len(r)-1)]-r[i])*(pos-i))
    totals={}
    paid=[x for x in rows if x['status']=='paid']
    for row in paid:
        k=row['region'];totals[k]=totals.get(k,Decimal(0))+Decimal(row['revenue'])
    d={'revenue_stats':{'count':len(r),'sum':float(sum(r)),'mean':float(statistics.mean(r)),
        'median':float(statistics.median(r)),'sample_stddev':float(statistics.stdev(r)),
        'p25':percentile('.25'),'p75':percentile('.75')},
       'region_totals':{k:float(v) for k,v in totals.items()},
       'largest_paid_ids':[int(x['id']) for x in sorted(paid,key=lambda x:(-Decimal(x['revenue']),int(x['id'])))[:5]],
       'weighted_unit_price':float(sum(r)/sum(Decimal(x['units']) for x in rows))}
    A=[[10,2,-1,0],[2,11,3,-1],[-1,3,12,2],[0,-1,2,9]]; b=[5,25,-2,19]
    # Exact Gauss-Jordan solution and independent permutation determinant.
    m=[[Fraction(x) for x in row]+[Fraction(rhs)] for row,rhs in zip(A,b)]
    for k in range(4):
        p=next(i for i in range(k,4) if m[i][k]);m[k],m[p]=m[p],m[k]
        q=m[k][k];m[k]=[x/q for x in m[k]]
        for i in range(4):
            if i!=k:
                q=m[i][k];m[i]=[x-q*y for x,y in zip(m[i],m[k])]
    determinant=sum((-1)**sum(p[i]>p[j] for i in range(4) for j in range(i+1,4))*math.prod(A[i][p[i]] for i in range(4)) for p in itertools.permutations(range(4)))
    lo,hi=0.,1.
    for _ in range(100):
        x=(lo+hi)/2
        if math.cos(x)>x:lo=x
        else:hi=x
    n={'solution':[float(row[-1]) for row in m],'determinant':determinant,
       'root':(lo+hi)/2,'integral':math.sqrt(math.pi)*math.erf(2)/2,
       'derivative':math.exp(.7)*(math.cos(.7)+math.sin(.7)),'tiny_root':1.25}
    return {'arithmetic':a,'data':d,'numerical':n}

def compare(want,got,path=''):
    if isinstance(want,dict):
        if not isinstance(got,dict):return [path+': expected object']
        return [error for k,v in want.items() for error in compare(v,got.get(k),path+'.'+k)]
    if isinstance(want,list):
        if not isinstance(got,list) or len(want)!=len(got):return [path+': array shape mismatch']
        return [error for i,(w,g) in enumerate(zip(want,got)) for error in compare(w,g,path+f'[{i}]')]
    if isinstance(want,(float,int)):
        tol=1e-7 if path.endswith('.derivative') else 1e-8 if path.endswith('.integral') else 1e-9 if path.endswith('root') else 1e-8
        if not isinstance(got,(float,int)) or not math.isclose(want,got,rel_tol=1e-12,abs_tol=tol):return [f'{path}: got {got!r}, expected {want!r}']
        return []
    return [] if want==got else [f'{path}: got {got!r}, expected {want!r}']

def stamp(t):return datetime.fromisoformat(t.replace('Z','+00:00'))

def collect(sessions,prefix):
    metrics={}
    for p in Path(sessions).glob('*.jsonl'):
        with p.open() as f:
            meta=json.loads(next(f))['payload'];name=meta.get('agent_path','')
            if not name.startswith('/root/'+prefix):continue
            events=[json.loads(x) for x in f]
        key=name.removeprefix('/root/'+prefix).replace('_','-')
        contexts=[e['payload'] for e in events if e['type']=='turn_context']
        usage=[e for e in events if e['type']=='token_usage_record']
        unique={e['payload']['response_id']:e for e in usage};usage=list(unique.values())
        sums={k:sum(e['payload']['usage'].get(k,0) for e in usage) for k in ['input_tokens','cached_input_tokens','cache_write_input_tokens','output_tokens','reasoning_output_tokens','total_tokens']}
        sums['uncached_input_tokens']=sums['input_tokens']-sums['cached_input_tokens']
        calls=[e for e in events if e['type']=='response_item' and e['payload'].get('type') in ('function_call','custom_tool_call')]
        starts=[e['timestamp'] for e in events if e['type']=='event_msg' and e['payload'].get('type')=='task_started']
        ends=[e['timestamp'] for e in events if e['type']=='event_msg' and e['payload'].get('type')=='task_complete']
        metrics[key]={'model':contexts[0].get('model'),'effort':contexts[0].get('effort'),
          'model_responses':len(usage),'outer_tool_calls':len(calls),
          'tool_call_names':[e['payload'].get('name') for e in calls],
          'generated_tool_argument_chars':sum(len(str(e['payload'].get('input',e['payload'].get('arguments','')))) for e in calls),
          'wall_seconds':(stamp(ends[-1])-stamp(starts[0])).total_seconds() if starts and ends else None,
          'tokens':sums,
          'usage_records':[{'timestamp':e['timestamp'],'usage':e['payload']['usage']} for e in usage]}
    return metrics

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--sessions');ap.add_argument('--prefix',default='bench_');ap.add_argument('--runs',default='runs');ap.add_argument('--output',default='initial-results.json');ap.add_argument('--repeat',type=int,default=30);args=ap.parse_args()
    expected=gold();results=collect(args.sessions,args.prefix) if args.sessions else {}
    names=[f'{s}-{a}' for s in SUITES for a in ARMS]
    for name in names:
        run=ROOT/args.runs/name;result=json.loads((run/'result.json').read_text())
        record=results.setdefault(name,{})
        record['correctness_errors']=compare(expected[name.rsplit('-',1)[0]],result)
        record['solution_bytes']=sum(p.stat().st_size for p in run.iterdir() if p.suffix in ('.sh','.py','.js') or p.name.startswith('query') and p.suffix=='.json')
        record['solution_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in run.iterdir() if p.is_file()}
        record['runtime_ms']=[]
    rng=random.Random(914)
    for i in range(args.repeat+3):
        rng.shuffle(names)
        for name in names:
            run=ROOT/args.runs/name;start=time.perf_counter_ns()
            proc=subprocess.run(['bash',str(run/'solution.sh')],capture_output=True,timeout=30)
            elapsed=(time.perf_counter_ns()-start)/1e6
            if proc.returncode:raise RuntimeError(f'{name}: {proc.stderr.decode()}')
            errors=compare(expected[name.rsplit('-',1)[0]],json.loads(proc.stdout))
            if errors:raise RuntimeError(f'{name}: {errors}')
            if i>=3:results[name]['runtime_ms'].append(elapsed)
    for r in results.values():
        if 'runtime_ms' not in r:continue
        xs=r['runtime_ms'];r['runtime_median_ms']=statistics.median(xs);r['runtime_p95_ms']=sorted(xs)[math.ceil(len(xs)*.95)-1]
    payload={'host':{'platform':platform.platform(),'machine':platform.machine(),'python':platform.python_version()},
      'repetitions':args.repeat,'warmups':3,'results':results,'gold':expected,
      'notes':['One fresh model run per arm per workload; no confidence interval for model metrics.',
      'Output tokens include reasoning; reasoning_output_tokens is a subset, not an additive category.',
      'Input tokens count all repeated context, including cache hits; uncached input is reported separately.',
      'Tool calls count outer orchestrator calls, which can contain multiple shell calls.',
      'Model tasks ran concurrently. Replay measured sequentially in seeded randomized order.',
      'No raw transcripts or hidden reasoning text are included.']}
    (ROOT/args.output).write_text(json.dumps(payload,indent=2)+'\n')
    for name,r in results.items():
        if 'runtime_ms' in r:print(name,'errors',len(r['correctness_errors']),'ms',round(r['runtime_median_ms'],2),'wall',r.get('wall_seconds'),'tokens',r.get('tokens',{}))
if __name__=='__main__':main()
