package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Regression: observing before the loader, consuming the one-shot body,
// trusting declared hashes, ignoring late changes or leaking arbitrary request
// fields would make the independent observer unsuitable for AC4/AC8.
func TestFlashClassIndependentObserver(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Fatal("Node is required for observer callback checks")
	}
	root := t.TempDir()
	module := filepath.Join(root, "node_modules", "@opencode", "plugin")
	if err := os.MkdirAll(module, 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"package.json": `{"type":"module","exports":"./index.js"}`, "index.js": `export const Plugin = { define: (value) => value }`} {
		if err := os.WriteFile(filepath.Join(module, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	source, err := os.ReadFile("../fixtures/flash-class-guidance/observer.js")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "pilot-observer"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"type":"module"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "pilot-observer", "index.js"), source, 0600); err != nil {
		t.Fatal(err)
	}
	script := `import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import { createHash } from 'node:crypto';
import plugin, { requestText, inspectReference } from './pilot-observer/index.js';
const canonical = '# Canonical reference\nDo not execute the reviewed plan.\n';
const resource = 'flow-plan/references/plan-review.md';
await fs.writeFile('canonical.md', canonical);
const hooks = new Map(); let disposed = 0;
const ctx = { options: {output:'proof.jsonl',canonical:'canonical.md',resource,role:'hive-review-plan'},
 plugin:{list:async()=>({data:[{id:'hive.context-loader',source:{type:'local'},state:{status:'active'}},{id:'hive.pilot-observer',source:{type:'local'},state:{status:'active'}}]})},
 agent:{ get:async ({agentID})=>({data:{id:agentID,name:agentID==='hive-review-plan'?'Plan Reviewer':'Build'},location:{}}) },
 session:{ get:async ({sessionID})=>({data:{id:sessionID,parentID:sessionID==='child'?'root':undefined},location:{}}), hook:async(name,callback)=>{hooks.set(name,callback);return {dispose:async()=>{disposed++}}} } };
const cleanup = await plugin.setup(ctx);
const model = {providerID:'opencode',id:'claude-haiku-5-5',variant:'high'};
const injected = 'Hive canonical reference: '+resource+'\n\n'+canonical;
const event = {sessionID:'child',agent:'hive-review-plan',model,system:[{type:'text',text:injected}],messages:[],options:{},tools:{read:{}}};
const before = JSON.stringify(event);
await hooks.get('context')(event);
assert.equal(JSON.stringify(event),before,'observer mutated assembled context');
await hooks.get('model.request')({...event,kind:'primary',headers:{authorization:'FAKE_PRIVATE_HEADER'}});
const request = new Request('https://example.invalid/private-token', {method:'POST',headers:{authorization:'FAKE_PRIVATE_HEADER'},body:JSON.stringify({messages:[{role:'system',content:injected}],private:'FAKE_PRIVATE_BODY'})});
await hooks.get('http.request')({...event,kind:'primary',request});
assert.equal(request.bodyUsed,false,'observer consumed native request');
const serialized = await fs.readFile('proof.jsonl','utf8');
for (const forbidden of [canonical,'FAKE_PRIVATE_HEADER','FAKE_PRIVATE_BODY','private-token','authorization']) assert.equal(serialized.includes(forbidden),false,'private text reached evidence');
let rows = serialized.trim().split('\n').map(JSON.parse);
assert.equal(rows[2].canonical,true);assert.equal(rows[2].copies,1);assert.equal(rows[2].correlated,true);
assert.equal(rows[2].hash,createHash('sha256').update(canonical).digest('hex'));
const changed = new Request('https://example.invalid/',{method:'POST',body:JSON.stringify({messages:[{role:'system',content:injected.replace('Do not','Do')} ]})});
await hooks.get('http.request')({...event,kind:'primary',request:changed});
const duplicated = new Request('https://example.invalid/',{method:'POST',body:JSON.stringify({messages:[{role:'system',content:injected},{role:'system',content:injected}]})});
await hooks.get('http.request')({...event,kind:'primary',request:duplicated});
await hooks.get('http.request')({...event,kind:'primary',request:new Request('https://example.invalid/',{method:'POST',body:'{"unknown":"FAKE_PRIVATE_BODY"}'})});
await hooks.get('experimental.ws.handshake')({...event,kind:'primary',url:'FAKE_PRIVATE_WS_URL',headers:{authorization:'FAKE_PRIVATE_HEADER'}});
await hooks.get('http.request')({...event,kind:'title'});
rows = (await fs.readFile('proof.jsonl','utf8')).trim().split('\n').map(JSON.parse);
assert.equal(rows[3].canonical,false);assert.equal(rows[3].correlated,false);
assert.equal(rows[4].copies,2);assert.equal(rows[4].correlated,false);
assert.equal(rows[5].code,'unsupported_protocol');assert.equal(rows[6].code,'unsupported_transport');
assert.equal(rows[7].phase,'http_auxiliary');
// Loader active but observer ordered before it: HTTP alone cannot repair the
// missing post-loader assembled-context snapshot.
const emptyContext = {...event,system:[]};
await hooks.get('context')(emptyContext);
await hooks.get('model.request')({...emptyContext,kind:'primary'});
await hooks.get('http.request')({...emptyContext,kind:'primary',request:new Request('https://example.invalid/',{method:'POST',body:JSON.stringify({messages:[{role:'system',content:injected}]})})});
rows = (await fs.readFile('proof.jsonl','utf8')).trim().split('\n').map(JSON.parse);
assert.equal(rows.at(-1).correlated,false,'observer ran before loader but claimed final coverage');
ctx.plugin.list = async()=>({data:[{id:'unexpected-later-plugin',source:{type:'local'},state:{status:'active'}}]});
await hooks.get('http.request')({...event,kind:'primary',request:new Request('https://example.invalid/',{method:'POST',body:JSON.stringify({messages:[{role:'system',content:injected}]})})});
rows = (await fs.readFile('proof.jsonl','utf8')).trim().split('\n').map(JSON.parse);
assert.equal(rows.at(-1).code,'plugin_inventory_unverified');
assert.equal(requestText({input:[{role:'system',content:[{type:'input_text',text:canonical}]}]}).protocol,'openai-responses');
assert.deepEqual(requestText({system:[],messages:[{role:'user',content:[{type:'tool_result',content:[{type:'text',text:canonical}]}]}]}).texts,[canonical]);
assert.deepEqual(requestText({contents:[{parts:[{text:canonical}]}]}).texts,[canonical]);
assert.equal(inspectReference([injected.replace('Do not','Do')],resource,canonical).canonical,false);
assert.equal(inspectReference([canonical,canonical],resource,canonical).copies,2);

// Bare responses, uncovered IDs and failed agent lookups preserve ID semantics.
ctx.plugin.list = async()=>({data:[{id:'hive.context-loader',source:{type:'local'},state:{status:'active'}},{id:'hive.pilot-observer',source:{type:'local'},state:{status:'active'}}]});
const readLast = async()=>JSON.parse((await fs.readFile('proof.jsonl','utf8')).trim().split('\n').at(-1));
assert.equal(rows[0].agent,'hive-review-plan');assert.equal(rows[0].agentName,'Plan Reviewer');assert.equal(rows[0].parentID,'root');assert.equal(rows[0].resource,resource);assert.equal(rows[0].copies,1);
ctx.agent.get = async()=>({id:'hive-review-plan',name:'Bare Label'});
ctx.session.get = async()=>({id:'child',parentID:'root'});
await hooks.get('context')(event);
let last = await readLast();assert.equal(last.resource,resource);assert.equal(last.copies,1);assert.equal(last.agentName,'Bare Label');assert.equal(last.parentID,'root');
ctx.agent.get = async()=>({data:{id:'other-role',name:'hive-review-plan'}});
await hooks.get('context')({...event,agent:'other-role'});
last = await readLast();assert.equal(last.resource,'');assert.equal(last.copies,0);assert.equal(last.agent,'other-role');
ctx.agent.get = async()=>{throw new Error('FAKE_PRIVATE_ERROR')};
await hooks.get('context')(event);
last = await readLast();assert.equal(last.agent,event.agent);assert.equal(last.resource,resource);assert.equal(last.copies,1);assert.equal(last.agentName,'');assert.equal(last.parentID,'root');
await hooks.get('model.request')({...event,kind:'primary'});
await hooks.get('http.request')({...event,kind:'primary',request:new Request('https://example.invalid/',{method:'POST',body:JSON.stringify({messages:[{role:'system',content:injected}]})})});
last = await readLast();assert.equal(last.correlated,true);assert.equal(last.copies,1);
assert.equal((await fs.readFile('proof.jsonl','utf8')).includes('FAKE_PRIVATE_ERROR'),false);
ctx.agent.get = ()=>{throw new Error('FAKE_PRIVATE_SYNC_ERROR')};
await hooks.get('context')(event);
last = await readLast();assert.equal(last.agent,event.agent);assert.equal(last.resource,resource);assert.equal(last.copies,1);
await cleanup();assert.equal(disposed,4);
`
	if err := os.WriteFile(filepath.Join(root, "check.mjs"), []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", "check.mjs")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("observer callback check failed: %v\n%s", err, output)
	}
}
