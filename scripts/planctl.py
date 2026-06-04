#!/usr/bin/env python3
"""Living-plan controller: keeps plan.json status/timestamps/summary consistent.

Usage:
  planctl.py start    <STORY>
  planctl.py code     <STORY> <TEST> <path-to-test-file>     # store source, status=written
  planctl.py green    <STORY> <TEST>                         # status=passing, stamp green_at/last_run
  planctl.py run      <STORY> <TEST> <passing|failing>       # stamp last_run + status
  planctl.py addtest  <STORY> <test-json-file>               # append a new test object
  planctl.py complete <STORY> <fix-solution-text-file>       # status=completed (+ green_test_at)
"""
import json, sys, datetime, os

PLAN = os.path.join(os.path.dirname(__file__), "..", "plan.json")

def now():
    return datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

def load():
    with open(PLAN) as f: return json.load(f)

def save(p):
    # ensure every story has the green_test_at field (idempotent migration)
    for s in p["stories"]:
        s.setdefault("green_test_at", None)
    p["summary"]["stories_completed"] = sum(1 for s in p["stories"] if s["status"]=="completed")
    p["summary"]["tests_green"] = sum(1 for s in p["stories"] for t in s["tests"] if t["status"]=="passing")
    p["summary"]["last_updated"] = now()
    with open(PLAN,"w") as f: json.dump(p,f,indent=2,ensure_ascii=False)

def story(p,sid):
    for s in p["stories"]:
        if s["id"]==sid: return s
    raise SystemExit(f"story {sid} not found")

def test(s,tid):
    for t in s["tests"]:
        if t["id"]==tid: return t
    raise SystemExit(f"test {tid} not found in {s['id']}")

def main():
    a=sys.argv[1:]
    if not a: raise SystemExit(__doc__)
    p=load(); cmd=a[0]
    if cmd=="start":
        s=story(p,a[1])
        if not s["started_at"]: s["started_at"]=now()
        s["status"]="in_progress"
        print(f"{s['id']} started_at={s['started_at']}")
    elif cmd=="code":
        s=story(p,a[1]); t=test(s,a[2])
        t["test_code"]=open(a[3]).read(); t["status"]="written"
        if a[3] not in t["test_files"]: t["test_files"].append(a[3])
        print(f"{t['id']} written ({len(t['test_code'])} chars from {a[3]})")
    elif cmd in ("green","run"):
        s=story(p,a[1]); t=test(s,a[2])
        status="passing" if cmd=="green" else a[3]
        t["status"]=status; t["last_run_at"]=now()
        if status=="passing" and not t["green_at"]: t["green_at"]=now()
        if all(x["status"]=="passing" for x in s["tests"]) and not s.get("green_test_at"):
            s["green_test_at"]=now(); print(f"  -> all {s['id']} tests green @ {s['green_test_at']}")
        print(f"{t['id']} {status} last_run={t['last_run_at']} green_at={t['green_at']}")
    elif cmd=="addtest":
        s=story(p,a[1]); obj=json.load(open(a[2]))
        s["tests"].append(obj); print(f"added {obj['id']} to {s['id']}")
    elif cmd=="complete":
        s=story(p,a[1]); s["status"]="completed"; s["completed_at"]=now()
        s["fix_solution"]=open(a[2]).read().strip()
        if not s.get("green_test_at") and s["tests"] and all(t["status"]=="passing" for t in s["tests"]):
            s["green_test_at"]=now()
        print(f"{s['id']} completed_at={s['completed_at']} green_test_at={s.get('green_test_at')}")
    else:
        raise SystemExit(__doc__)
    save(p)

if __name__=="__main__": main()
