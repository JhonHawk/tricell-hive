#!/usr/bin/env python3
"""Regenerate the pinned catalog from the read-only 16e7d335 source checkout."""
import base64, hashlib, json, pathlib, subprocess, sys
r=pathlib.Path(sys.argv[1]); dest=pathlib.Path(__file__).parent
revision='16e7d3357a3c41530d5e31460c3024872566f3c7'
assert subprocess.check_output(['git','-C',str(r),'rev-parse','HEAD'],text=True).strip()==revision
files={}
def add(root,path,src,manifest=''):
 if not src.is_file(): return
 data=subprocess.check_output(['git','-C',str(r),'show',revision+':'+str(src.relative_to(r))]);files[(root,path)]={'root':root,'path':path,'source':str(src.relative_to(r)),'sha256':hashlib.sha256(data).hexdigest(),'data':base64.b64encode(data).decode(),'manifest':manifest}
def tree(root,base,src,mp='',flat=False):
 for p in sorted(src.rglob('*')):
  if p.is_file() and '__pycache__' not in p.parts and p.suffix!='.pyc':
   rel=p.name if flat else p.relative_to(src).as_posix();add(root,base+rel,p,mp+rel)
add('claude','CLAUDE.md',r/'global/CLAUDE.md','CLAUDE.md')
tree('claude','skills/',r/'global/skills','skills/')
for p in (r/'harness/agents-skills').glob('*/references/*'):
 if p.is_file() and not (r/'global/skills'/p.relative_to(r/'harness/agents-skills')).exists():add('claude','skills/'+p.relative_to(r/'harness/agents-skills').as_posix(),p,'skills/'+p.relative_to(r/'harness/agents-skills').as_posix())
tree('claude','agents/',r/'harness/claude/agents','agents/')
tree('shared','skills/',r/'harness/agents-skills','agents-skills/')
for host in ['codex','opencode','grok']:
 tree(host,'agents/',r/f'harness/{host}/agents',f'{host}-agents/',True)
 files.pop((host,'agents/README.md'),None)
 if host!='grok':add(host,'AGENTS.md',r/'harness/AGENTS.md',f'harness-agents/{host}/AGENTS.md')
tree('opencode','commands/',r/'harness/opencode/commands','opencode-commands/',True)
for p in (r/'global/hooks').rglob('*'):
 if not p.is_file() or p.name.startswith('test_') or p.name.endswith('.test.ts'):continue
 if p.suffix in ['.sh','.py'] or (p.suffix=='.json' and p.name not in ['settings-config.json','codex-hooks.json']):add('claude','hooks/'+p.name,p,'hooks/'+p.name)
 if p.suffix in ['.sh','.py'] and (p.parent/'codex-hooks.json').exists():add('codex','hooks/'+p.name,p,'codex-hooks/'+p.name)
 if p.suffix=='.ts':add('opencode','plugins/'+p.name,p,'opencode-plugins/'+p.name)
for h in ['claude','codex']:add(h,'hooks/rule-manifest.json',r/'harness/rule-manifest.json',('hooks/' if h=='claude' else 'codex-hooks/')+'rule-manifest.json')
for d in ['agents','src','extensions']:tree('pi',d+'/',r/f'harness/pi/{d}')
files.pop(('pi','agents/README.md'),None)
add('pi','AGENTS.md',r/'harness/AGENTS.md')
for x in ['bash-policy/bash-policy.sh','flow-context/flow-context.sh','flow-session-context/flow-session-context.sh','post-tool-hub/post-tool-hub.sh','reviewer-guard/reviewer-guard.sh','rule-delivery/rule-delivery.py','session-hygiene-report/session-hygiene-report.sh']:add('pi','global/hooks/'+x,r/'global/hooks'/x)
add('pi','global/hooks/rule-delivery/rule-manifest.json',r/'harness/rule-manifest.json')
hooks=[]
for p in (r/'global/hooks').rglob('*.json'):
 if p.name not in ['settings-config.json','codex-hooks.json']:continue
 obj=json.loads(p.read_text());host='codex' if p.name=='codex-hooks.json' else 'claude'
 for ev,entries in obj.get('hooks',{}).items():
  for entry in entries:
   for hook in entry.get('hooks',[]):
    if 'command' in hook:hooks.append({'host':host,'event':ev,'command':hook['command']})
obj={'revision':revision,'files':list(files.values()),'hooks':hooks}
(dest/'catalog.json').write_text(json.dumps(obj,separators=(',',':'))+'\n')
for f in ['pi-subagents-0.67.0.json','pi-subagents-0.67.0.patch']:(dest/f).write_bytes((r/'harness/pi/patches'/f).read_bytes())
print(len(files),'files;',len(hooks),'hooks')
