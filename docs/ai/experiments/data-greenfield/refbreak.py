#!/usr/bin/env python3
"""Rates over a set of kinds, split by whether a missing path is a reference.

A reference path is any path carrying a <thing>Ref or <thing>Refs segment,
optionally an array. That covers the object, its .external/.name/.namespace/.kind
children, and the repeated form -- each unreproduced reference costs four or five
missing paths, so counting them as ordinary fields overstates the gap fourfold.
"""
import os, re, sys, glob

REF = re.compile(r'(^|\.)[A-Za-z0-9_]+Refs?(\[\])?(\.|$)')

def analyse(verbose_dir, kinds):
    tot = {'spec': [0,0,0], 'required': [0,0,0], 'status.observedState': [0,0,0]}
    miss = {'spec': {'ref': 0, 'other': []}, 'status.observedState': {'ref': 0, 'other': []}}
    refs = [0,0,0,0]   # baseline, reproduced, plain, absent
    n = 0
    for kind in kinds:
        f = os.path.join(verbose_dir, kind + '.txt')
        if not os.path.exists(f):
            continue
        n += 1
        sec = None
        for line in open(f):
            m = re.match(r'\s+(spec|required|status\.observedState)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s', line)
            if m:
                b = m.group(1)
                tot[b][0] += int(m.group(2)); tot[b][1] += int(m.group(3)); tot[b][2] += int(m.group(5))
                continue
            m = re.search(r'references: (\d+) in baseline -> (\d+) reproduced, (\d+) left as plain strings, (\d+) absent', line)
            if m:
                for i in range(4): refs[i] += int(m.group(i+1))
                continue
            m = re.match(r'\s+(spec|status\.observedState) missing \((\d+)\):', line)
            if m: sec = m.group(1); continue
            if re.match(r'\s+\S+ (missing|extra|mismatch)', line): sec = None; continue
            if sec and line.startswith('      '):
                p = line.strip().split(' ')[0]
                if REF.search(p): miss[sec]['ref'] += 1
                else: miss[sec]['other'].append((kind, p))
    return tot, miss, refs, n

def rate(m, mi, mm):
    d = m + mi + mm
    return 100.0 * m / d if d else 100.0

if __name__ == '__main__':
    vdir, kindfile, label = sys.argv[1], sys.argv[2], sys.argv[3]
    kinds = [l.strip() for l in open(kindfile) if l.strip()]
    tot, miss, refs, n = analyse(vdir, kinds)
    print(f"{label}  (n={n})")
    for b in ('spec', 'required', 'status.observedState'):
        m, mi, mm = tot[b]
        line = f"  {b:22s} {rate(m,mi,mm):5.1f}%"
        if b in miss:
            nonref = len(miss[b]['other'])
            line += f"   excl refs {rate(m,nonref,mm):5.1f}%   missing {mi} = {miss[b]['ref']} ref + {nonref} other"
        print(line)
    print(f"  references: {refs[0]} baseline -> {refs[1]} reproduced, {refs[2]} plain string, {refs[3]} absent")
