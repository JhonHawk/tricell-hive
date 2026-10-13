package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Regression: correct helpers are insufficient if main skips the observer,
// accepts a parent's finish instead of its child, loses memory provenance or
// leaves an imported credential copy. Exercise the actual runner with fake CLIs.
func TestFlashClassRunnerWithFakeNativeCLIs(t *testing.T) {
	if os.Getenv("FLASH_RUNNER_SUBPROCESS") == "1" {
		flag.CommandLine = flag.NewFlagSet("pilot", flag.ExitOnError)
		os.Args = append([]string{"pilot"}, strings.Split(os.Getenv("FLASH_RUNNER_ARGS"), "\n")...)
		main()
		return
	}
	for _, mode := range []string{"complete", "unfinished child", "parent external", "child transient", "child opaque", "child final"} {
		t.Run(mode, func(t *testing.T) {
			bin, home, outputParent := t.TempDir(), t.TempDir(), t.TempDir()
			source := newFixtureCheckout(t, "# Fake guidance\n")
			write := func(name, body string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cases, err := os.ReadFile("../fixtures/flash-class-guidance/cases.json")
			if err != nil {
				t.Fatal(err)
			}
			observer, err := os.ReadFile("../fixtures/flash-class-guidance/observer.js")
			if err != nil {
				t.Fatal(err)
			}
			write(filepath.Join(source, "tests/fixtures", flashSuite, "cases.json"), string(cases))
			write(filepath.Join(source, "tests/fixtures", flashSuite, "observer.js"), string(observer))
			profiles, err := os.ReadFile("../../integrations/agent-profiles.json")
			if err != nil {
				t.Fatal(err)
			}
			write(filepath.Join(source, "integrations/agent-profiles.json"), string(profiles))
			write(filepath.Join(source, "content/agents/review/hive-review-plan.md"), "---\nname: hive-review-plan\ndescription: Review the plan.\nmodel_profile: reasoning\naccess_profile: observe\neffort: medium\n---\n# Keep role instructions\n")
			write(filepath.Join(source, "content/skills/flow-plan/SKILL.md"), "---\nname: flow-plan\ndescription: Plan authorized work.\n---\n# Plan\n")
			write(filepath.Join(source, "content/skills/flow-plan/references/plan-review.md"), "# Review\nRead only.\n")
			write(filepath.Join(source, "content/skills/harness-audit/SKILL.md"), "---\nname: harness-audit\ndescription: Audit instructions.\n---\n# Audit\n")
			write(filepath.Join(source, "integrations/opencode/context-loader.js"), "const snapshot = __HIVE_CONTEXT_SNAPSHOT_DATA__;\nexport default {};\n")
			write(filepath.Join(home, ".config/opencode/opencode.json"), `{"mcp":{"engram":{"type":"local","enabled":true,"command":["engram","mcp"]}}}`)
			write(filepath.Join(home, ".cache/opencode/models.json"), "{}")
			engram := `#!/usr/bin/env python3
import http.server,json,os,pathlib,sys
a=sys.argv[1:];store=pathlib.Path(os.environ['ENGRAM_DATA_DIR']);records=store/'records.json'
if a==['projects','list']: print('No projects found.');sys.exit(0)
if a==['instance-id']: print('0123456789abcdef0123456789abcdef');sys.exit(0)
if a[0]=='serve':
 class Handler(http.server.BaseHTTPRequestHandler):
  def do_GET(self):
   self.send_response(200);self.end_headers();self.wfile.write(b'{"instance_id":"0123456789abcdef0123456789abcdef"}')
  def log_message(self,*_): pass
 http.server.HTTPServer(('127.0.0.1',int(a[1])),Handler).serve_forever()
if a[0]=='stats': print('Sessions: 1\nObservations: 1\nPrompts: 0');sys.exit(0)
doc=json.loads(records.read_text()) if records.exists() else {'version':'3','sessions':[],'observations':[],'prompts':[]}
if a[0]=='export': pathlib.Path(a[1]).write_text(json.dumps(doc));sys.exit(0)
if a[0]=='delete':
 if a[1]=='session': key,id='sessions',a[2]
 elif a[1]=='prompt': key,id='prompts',int(a[2])
 else: key,id='observations',int(a[1])
 doc[key]=[x for x in doc[key] if x['id']!=id];records.write_text(json.dumps(doc));sys.exit(0)
sys.exit(1)
`
			opencode := `#!/usr/bin/env python3
import hashlib,json,os,pathlib,sys
a=sys.argv[1:];home=pathlib.Path(os.environ['HOME'])
if a==['--version']: print('opencode v2.0.26 fake');sys.exit(0)
if a[:2]==['auth','export']: print('{"marker":"FAKE_CREDENTIAL"}');sys.exit(0)
if a[:2]==['auth','import']:
 data=home/'.local/share/opencode';data.mkdir(parents=True,exist_ok=True);(data/'opencode.db').write_text(sys.stdin.read());sys.exit(0)
cfg=json.loads((home/'.config/opencode/opencode.json').read_text());options=cfg['plugins'][-1]['options'];model=cfg['agents']['hive-review-plan']['model'];provider,rest=model.split('/',1);id,variant=rest.split('#',1);ref={'providerID':provider,'id':id,'variant':variant}
if a[0]=='run':
 if os.environ.get('FAKE_OPERATION')=='parent external':
  p=pathlib.Path(os.environ['FAKE_EXTERNAL_PATH']);original=p.read_bytes();p.write_text('transient parent write');p.write_bytes(original)
  print(json.dumps({'type':'tool_use','sessionID':'root','part':{'tool':'write','callID':'parent-write','messageID':'review','state':{'input':{'filePath':str(p)},'status':'completed','output':'done'}}}))
 project=cfg['mcp']['servers']['engram']['environment']['ENGRAM_PROJECT'];store=pathlib.Path(os.environ['ENGRAM_DATA_DIR']);(store/'records.json').write_text(json.dumps({'version':'3','sessions':[{'id':'memory-child','project':project}],'observations':[{'id':101,'project':project}],'prompts':[]}))
 canonical=pathlib.Path(options['canonical']).read_bytes();hash=hashlib.sha256(canonical).hexdigest()
 rows=[{'phase':'http','sessionID':'root','parentID':'','agent':'build','agentName':'build','model':model,'sequence':1,'resource':'','hash':'','copies':0,'canonical':False,'correlated':True,'protocol':'chat-completions','kind':'primary'}, {'phase':'context','sessionID':'child','parentID':'root','agent':'hive-review-plan','agentName':'Plan Reviewer','model':model,'sequence':1,'resource':options['resource'],'hash':hash,'copies':1,'injectedCopies':1,'canonical':True,'correlated':False,'protocol':'assembled','kind':'primary'}, {'phase':'http','sessionID':'child','parentID':'root','agent':'hive-review-plan','agentName':'Plan Reviewer','model':model,'sequence':1,'resource':options['resource'],'hash':hash,'copies':1,'injectedCopies':1,'canonical':True,'correlated':True,'protocol':'chat-completions','kind':'primary'}]
 for row in rows: row['loaderActive']=True
 pathlib.Path(options['output']).write_text(''.join(json.dumps(row)+'\n' for row in rows))
 for event in [{'type':'tool_use','sessionID':'root','part':{'tool':'subagent','callID':'dispatch','messageID':'review','state':{'input':{'agent':'hive-review-plan'},'status':'completed','output':'Review completed.'}}},{'type':'step_finish','sessionID':'root','part':{'reason':'stop','messageID':'final'}}]: print(json.dumps(event))
 sys.exit(0)
if a[:2]==['session','export']:
 session=a[-1];primary=session=='root';agent='build' if primary else 'hive-review-plan';stop=primary or os.environ.get('FAKE_CHILD_UNFINISHED')!='1'
 content=[{'type':'tool','id':'dispatch','name':'subagent','state':{'status':'completed'}}] if primary else [{'type':'text','text':'Unique child finding: AC2 has no revoked-account task.'}]
 if not primary and os.environ.get('FAKE_OPERATION')=='child transient':
  p=pathlib.Path(os.environ['FAKE_EXTERNAL_PATH']);original=p.read_bytes();p.write_text('transient child write');p.write_bytes(original)
  content.append({'type':'tool','id':'child-write','name':'write','state':{'status':'completed','input':{'filePath':str(p),'private':'FAKE_PRIVATE_TOOL_ARG'},'content':[{'type':'text','text':'FAKE_PRIVATE_TOOL_RESULT'}]},'providerState':{'secret':'FAKE_PRIVATE_PROVIDER_STATE'}})
 if not primary and os.environ.get('FAKE_OPERATION')=='child opaque': content.append({'type':'tool','id':'child-shell','name':'bash','state':{'status':'completed','input':{'command':'OPAQUE_FAKE_PRIVATE_COMMAND'},'content':[]}})
 print(json.dumps({'info':{'id':session,'parentID':'' if primary else 'root','agent':agent,'model':ref,'outcome':'succeeded' if stop else 'interrupted','time':{'idle':1}},'messages':[{'id':'final' if primary else 'child-final','type':'assistant','agent':agent,'model':ref,'finish':'stop' if stop else 'tool-calls','time':{'completed':1},'content':content}]}));sys.exit(0)
sys.exit(1)
`
			for name, body := range map[string]string{"engram": engram, "opencode": opencode} {
				path := filepath.Join(bin, name)
				write(path, body)
				if err := os.Chmod(path, 0700); err != nil {
					t.Fatal(err)
				}
			}
			out := filepath.Join(outputParent, "run")
			args := []string{"--suite", flashSuite, "--host", "opencode", "--case", "plan-reference", "--arm", "C", "--guidance-source", source, "--source", source, "--model", flashModels[0], "--out", out, "--timeout", "3m"}
			cmd := exec.Command(os.Args[0], "-test.run=^TestFlashClassRunnerWithFakeNativeCLIs$")
			cmd.Env = append(os.Environ(), "FLASH_RUNNER_SUBPROCESS=1", "FLASH_RUNNER_ARGS="+strings.Join(args, "\n"), "HOME="+home, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "ENGRAM_DATA_DIR="+filepath.Join(home, "everyday"), "XDG_CONFIG_HOME=", "XDG_CACHE_HOME=")
			if mode == "unfinished child" {
				cmd.Env = append(cmd.Env, "FAKE_CHILD_UNFINISHED=1")
			}
			victim := filepath.Join(t.TempDir(), "external.txt")
			write(victim, "Keep external fixture unchanged.\n")
			cmd.Env = append(cmd.Env, "FAKE_OPERATION="+mode, "FAKE_EXTERNAL_PATH="+victim)
			if stdout, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("fake runner failed: %v\n%s", err, stdout)
			}
			data, err := os.ReadFile(filepath.Join(out, "run.json"))
			if err != nil {
				t.Fatal(err)
			}
			var r result
			if json.Unmarshal(data, &r) != nil {
				t.Fatal("runner report malformed")
			}
			if r.FlashClass == nil || r.FlashClass.Memory != "blocked" || !r.GuidanceVariant.ShadowCredentialStoreRemoved {
				t.Fatal("runner failed to withhold unbound memory cleanup or remove the stopped credential copy")
			}
			if mode == "complete" && (r.Terminal == "completed" || !strings.Contains(r.FlashClass.Error, "atomic memory store binding unavailable") || r.FlashClass.Completion != "observed") {
				t.Fatal("completed native review incorrectly waived the missing memory cleanup contract")
			}
			if mode == "unfinished child" && r.Terminal == "completed" {
				t.Fatal("parent exit closed an unfinished child")
			}
			if strings.Contains(mode, "external") || strings.Contains(mode, "transient") || strings.Contains(mode, "opaque") {
				if r.Terminal == "completed" {
					t.Fatal("observed unsafe operation allowed batch continuation")
				}
			}
			if mode == "child final" || mode == "child transient" || mode == "child opaque" {
				entries, err := filepath.Glob(filepath.Join(out, "native-review-*.json"))
				if err != nil || len(entries) == 0 {
					t.Fatal("native child review evidence discarded before cleanup")
				}
				found := false
				for _, entry := range entries {
					proof, err := os.ReadFile(entry)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Contains(string(proof), "Unique child finding: AC2") {
						found = true
					}
					for _, secret := range []string{"FAKE_PRIVATE_TOOL_ARG", "FAKE_PRIVATE_TOOL_RESULT", "FAKE_PRIVATE_PROVIDER_STATE", "OPAQUE_FAKE_PRIVATE_COMMAND"} {
						if strings.Contains(string(proof), secret) {
							t.Fatal("unnecessary native private fields persisted")
						}
					}
					info, err := os.Stat(entry)
					if err != nil || info.Mode().Perm() != 0600 {
						t.Fatal("native projection is not private")
					}
				}
				if !found {
					t.Fatal("parent summary replaced the child's actual final review")
				}
			}
			kept, err := os.ReadFile(victim)
			if err != nil || string(kept) != "Keep external fixture unchanged.\n" {
				t.Fatal("controlled victim was not restored")
			}
			if _, err := os.Stat(filepath.Join(out, "engram-data")); err != nil || r.MemoryIsolation.Cleanup != "cleanup_withheld_unverified" {
				t.Fatal("uncertain memory store was deleted instead of retained and reported")
			}
			if _, err := os.Stat(filepath.Join(out, "shadow-home/.local")); !os.IsNotExist(err) {
				t.Fatal("native credential/history directory survived cleanup")
			}
		})
	}
}
