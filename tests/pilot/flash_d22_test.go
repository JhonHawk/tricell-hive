package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFlashD22SimpleCommandAllowlist(t *testing.T) {
	denied := []string{`mysql -pdemo -e "SELECT 1"`, `SSHPASS=demo sshpass -e ssh example.invalid true`, "cat file | wc", "cat file > output", "ls; pwd", "cat $(pwd)", "ENV=value ls", "find . -exec cat file", "find . -execdir ls", "find . -delete", "find . -fprint output", "find . -fprintf output", "git push", "git commit", "ls *.go", "cat ../file &", "pwd\nls", "ls `pwd`", "ls ${HOME}", "cat < file"}
	for i, command := range denied {
		if flashOperationCommand(command) != "" {
			t.Errorf("unrecognized command accepted at index %d", i)
		}
	}
	for _, command := range []string{"ls", "cat /fixture/plan.md", "rg needle /fixture", "git status", "head file", "tail file", "wc file", "pwd", "grep needle file", "find /fixture -type f", "git log", "git diff", "git show", "git rev-parse HEAD"} {
		if flashOperationCommand(command) != command {
			t.Error("simple read-only command omitted")
		}
	}
}

func TestFlashD22RoleRejectsEscapingAgentsDirectory(t *testing.T) {
	home := t.TempDir()
	config := filepath.Join(home, ".config/opencode")
	if err := os.MkdirAll(config, 0700); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	target := filepath.Join(outside, "hive-review-plan.md")
	raw := "---\nmodel: old/provider\n---\nKeep outside bytes\n"
	if err := os.WriteFile(target, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(config, "agents")); err != nil {
		t.Fatal(err)
	}
	// The root stays at the shadow home, above every configurable ancestor.
	if _, _, _, err := overrideFlashRoleModel(home, "hive-review-plan", flashModels[0]); err == nil {
		t.Error("escaping agents directory accepted")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != raw {
		t.Error("outside role changed")
	}
}

func TestFlashD22CatalogSnapshotHash(t *testing.T) {
	script := `import importlib.util,json,hashlib,sys
spec=importlib.util.spec_from_file_location('launcher',sys.argv[1]);m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
obj=m.Launcher.__new__(m.Launcher);obj.data={};obj.container_id=lambda:'fake-container';obj.save=lambda:None
expected=hashlib.sha256(b'fake snapshot bytes').hexdigest()
def call(*args,**kwargs):
 if 'opencode' in args: return '\n'.join('opencode-go/'+model for model in m.MODELS)
 if 'cat' in args: return json.dumps({'opencode-go':{'models':{model:{} for model in m.MODELS}}})
 if 'sha256sum' in args: return expected+'  /home/pilot/.cache/opencode/models.json\n'
 return ''
obj.call=call;obj.catalog()
assert obj.data.get('models_snapshot_sha256')==expected,'catalog snapshot hash missing'
`
	cmd := exec.Command("python3", "-I", "-c", script, "container/launcher.py")
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Fatal("fake catalog manifest lacks snapshot hash")
	}
}
