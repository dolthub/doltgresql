#!/usr/bin/env python3
"""Discover and run the complete server suite, isolating each top-level test.
Build go.test, engine.test and extensions.test first. Excludes the regression
replay and non-server unit tests; does not use a feature allowlist.
"""
import argparse,concurrent.futures,json,os,pathlib,re,signal,subprocess,threading,time
SRC=pathlib.Path(__file__).resolve().parents[2]
UNIT_NAMES={'TestParserBehaviors','TestLegacyJsonBRoundTrip','TestLegacyJsonBReserialization','TestScriptErrorDetail','TestJsonbOutCallableWithAnyWrapper','TestJsonOutCallableWithAnyWrapper','TestJsonbSerializeWithAnyWrapper','TestDropRoleCleansPersistedAuthorizationReferences','TestExternalServerCleanup','TestTargetExpectationSelection'}
PACKAGES={'go':('testing/go','go.test'),'enginetest':('testing/go/enginetest','engine.test'),'extensions':('testing/go/extensions','extensions.test')}
LOCK=threading.Lock()

def inventory(build,env):
    jobs=[]
    for package,(directory,binary) in PACKAGES.items():
        files={}
        for f in (SRC/directory).glob('*_test.go'):
            for name in re.findall(r'^func (Test\w+)\(t \*testing.T\)',f.read_text(),re.M): files[name]=str(f.relative_to(SRC))
        names=subprocess.check_output([str(build/binary),'-test.list','^Test'],cwd=SRC/directory,env=env,text=True).splitlines()
        for name in names:
            if not name.startswith('Test'):continue
            source=files.get(name,'')
            if not source:raise RuntimeError('Unmapped test '+name)
            exclusion=None
            if name=='TestRegressions':exclusion='requested regression replay exclusion'
            elif name in UNIT_NAMES or source.endswith('/query_converter_test.go'):exclusion='non-server unit test'
            jobs.append(dict(package=package,name=name,source=source,directory=directory,binary=str(build/binary),exclude=exclusion))
    return jobs

def run(job,target,env,root,timeout,attempt='initial'):
    name=job['name'];folder=root/'runs'/attempt/target/job['package'];folder.mkdir(parents=True,exist_ok=True)
    events=folder/(name+'.jsonl');log=folder/(name+'.log');start=time.monotonic()
    command=['go','tool','test2json','-t','-p',job['package'],job['binary'],'-test.run','^'+re.escape(name)+'$','-test.v=test2json','-test.count=1',f'-test.timeout={timeout}s']
    with events.open('w') as out:
        p=subprocess.Popen(command,cwd=SRC/job['directory'],env=env,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
        timedout=False
        try:rc=p.wait(timeout=timeout+15)
        except subprocess.TimeoutExpired:timedout=True;rc=-1
        finally:
            # SIGINT asks PostgreSQL for fast shutdown, unlike SIGTERM's smart shutdown.
            # This group belongs only to this test; never touch another worker's server.
            try:os.killpg(p.pid,signal.SIGINT)
            except ProcessLookupError:pass
            if p.poll() is None:
                try:p.wait(timeout=5)
                except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
    statuses={};output=[]
    for line in events.read_text(errors='replace').splitlines():
        try:e=json.loads(line)
        except ValueError:output.append(line+'\n');continue
        if e.get('Output'):output.append(e['Output'])
        if e.get('Test') and e.get('Action') in ['pass','fail','skip']:statuses[e['Test']]=e['Action']
    log.write_text(''.join(output))
    r=dict(job,target=target,attempt=attempt,rc=rc,seconds=round(time.monotonic()-start,2),status='timeout' if timedout else statuses.get(name,'aborted' if rc else 'unknown'),statuses=statuses,log=str(log.relative_to(root)),events=str(events.relative_to(root)))
    with LOCK:
        with (root/'results.jsonl').open('a') as out:out.write(json.dumps(r)+'\n')
        print(target,job['package'],name,r['status'],r['seconds'],flush=True)
    return r

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--build-dir',type=pathlib.Path,required=True);p.add_argument('--output',type=pathlib.Path,required=True)
    p.add_argument('--stock',required=True);p.add_argument('--fork',required=True);p.add_argument('--doltgres',required=True)
    p.add_argument('--workers',type=int,default=3);p.add_argument('--timeout',type=int,default=1200)
    p.add_argument('--targets',nargs='+',choices=['doltgres','postgres','postgres-fork'],default=['doltgres','postgres','postgres-fork'])
    p.add_argument('--attempt',default='initial');p.add_argument('--tests',nargs='*',help='Diagnostic reruns only; omit for a full sweep')
    a=p.parse_args();a.output.mkdir(parents=True,exist_ok=True)
    env={k:v for k,v in os.environ.items() if k not in ['OPENAI_API_KEY','ANTHROPIC_OAUTH_TOKEN','GITHUB_TOKEN','REPO_ACCESS_TOKEN','DOLTGRES_TEST_PROFILE','PGOPTIONS','GITHUB_ACTION','DOLTGRES_TEST_STATEMENT_TIMEOUT']}
    env.update(LC_ALL='C',GOMAXPROCS='2',DOLTGRES_TEST_RECOVER_PANICS='1')
    jobs=inventory(a.build_dir,env);(a.output/'inventory.json').write_text(json.dumps(jobs,indent=2))
    environments={}
    for target in a.targets:
        e=dict(env,DOLTGRES_TEST_TARGET=target,DOLTGRES_TEST_SERVER_KIND='doltgres' if target=='doltgres' else 'postgres',DOLTGRES_TEST_SERVER_BINARY={'doltgres':a.doltgres,'postgres':a.stock,'postgres-fork':a.fork}[target],DOLTGRES_TEST_TABLE_AM='heap' if target=='postgres' else 'dolt_prolly_table')
        environments[target]=e
    done={(r['attempt'],r['target'],r['package'],r['name']) for r in map(json.loads,(a.output/'results.jsonl').read_text().splitlines())} if (a.output/'results.jsonl').exists() else set()
    todo=[(j,t) for j in jobs for t in a.targets if not j['exclude'] and (not a.tests or j['name'] in a.tests) and (a.attempt,t,j['package'],j['name']) not in done]
    print('Discovered',len(jobs),'eligible',sum(not j['exclude'] for j in jobs),'target runs queued',len(todo),flush=True)
    # All targets share fixed engine ports. Serialize this package across targets.
    def engines():
        for j,t in todo:
            if j['package']=='enginetest':run(j,t,environments[t],a.output,a.timeout,a.attempt)
    with concurrent.futures.ThreadPoolExecutor(max_workers=a.workers) as pool:
        fs=[pool.submit(engines)]
        fs += [pool.submit(run,j,t,environments[t],a.output,a.timeout,a.attempt) for j,t in todo if j['package']!='enginetest']
        for future in concurrent.futures.as_completed(fs):future.result()
    print('Sweep complete',flush=True)

if __name__=='__main__':main()
