package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/version"
)

func runReleaseExecutor(in io.Reader, out io.Writer) error {
	q, err := distribution.ReadUpdateRequest(in)
	if err != nil {
		return err
	}
	response := distribution.UpdateResponse{Protocol: q.Protocol, OperationID: q.OperationID, Action: q.Action, ProductVersion: q.ProductVersion}
	result := management.UpdateResultPayload{}
	var binding distribution.UpdatePackageBinding
	var plan management.UpdatePlan
	var prepare management.UpdatePreparePayload
	var recoverPayload management.UpdateRecoverPayload
	switch q.Action {
	case distribution.UpdatePrepare:
		err = distribution.DecodeUpdateMessage(bytes.NewReader(q.Payload), &prepare)
		binding = prepare.Package
	case distribution.UpdateApply:
		var payload management.UpdateApplyPayload
		err = distribution.DecodeUpdateMessage(bytes.NewReader(q.Payload), &payload)
		plan = payload.Plan
		binding = plan.Package
	case distribution.UpdateRecover:
		err = distribution.DecodeUpdateMessage(bytes.NewReader(q.Payload), &recoverPayload)
		plan = recoverPayload.Journal.Plan
		binding = plan.Package
	}
	if err == nil && (q.ProductVersion != version.Current || binding.Artifact.ProductVersion != q.ProductVersion) {
		err = fmt.Errorf("target executor product version mismatch")
	}
	if err == nil {
		executable, e := os.Executable()
		if e == nil {
			executable, e = filepath.EvalSymlinks(executable)
		}
		expected := binding.TargetExecutable.Path
		if q.Action == distribution.UpdatePrepare {
			expected = binding.PreparationExecutor.Path
		}
		if e != nil || executable != expected {
			err = fmt.Errorf("update invoked outside its bound executor")
		}
	}
	if err == nil {
		switch q.Action {
		case distribution.UpdatePrepare:
			var p management.UpdatePlan
			p, err = management.PrepareReleaseUpdate(prepare, q.OperationID)
			if err == nil {
				result.Plan = &p
				response.Status = distribution.UpdatePrepared
			}
		case distribution.UpdateApply:
			if plan.Installation {
				err = fmt.Errorf("package installation must be confirmed through hive install")
				break
			}
			if q.OperationID != plan.OperationID {
				err = fmt.Errorf("update operation mismatch")
				break
			}
			var phase management.UpdatePhase
			phase, err = (management.Engine{}).ApplyReleaseUpdate(plan)
			response.Status = distribution.UpdateStatus(phase)
		case distribution.UpdateRecover:
			if err = recoverPayload.Journal.Validate(); err != nil {
				break
			}
			if q.OperationID != plan.OperationID {
				err = fmt.Errorf("update operation mismatch")
				break
			}
			current, e := management.PendingUpdate(plan.StateDir)
			if e != nil {
				err = e
				break
			}
			if current == nil || current.Plan.ID != plan.ID {
				err = fmt.Errorf("pending update identity mismatch")
				break
			}
			var phase management.UpdatePhase
			phase, err = (management.Engine{}).RecoverReleaseUpdate(plan.StateDir)
			response.Status = distribution.UpdateStatus(phase)
		}
	}
	if err != nil {
		response.Status = distribution.UpdateFailed
		response.Error = err.Error()
	} else {
		result.CoreTransactionID = plan.CoreTransactionID
		if plan.CoreTransactionID != "" {
			result.CoreJournalPath = filepath.Join(plan.StateDir, "transactions", plan.CoreTransactionID+".json")
		}
		response.Payload, _ = json.Marshal(result)
	}
	return distribution.WriteUpdateMessage(out, response)
}

func routeSavedUpdate(data []byte) (bool, error) {
	if !management.IsUpdatePlan(data) {
		return false, nil
	}
	p, err := management.ReadUpdatePlan(bytes.NewReader(data))
	if err != nil {
		return true, err
	}
	if p.Installation {
		return true, fmt.Errorf("package installation must be confirmed through hive install")
	}
	payload, _ := json.Marshal(management.UpdateApplyPayload{Plan: p})
	q := distribution.UpdateRequest{Protocol: p.Protocol, OperationID: p.OperationID, Action: distribution.UpdateApply, ProductVersion: p.Package.Artifact.ProductVersion, Payload: payload}
	r, err := management.RouteUpdate(q, p)
	if err != nil {
		return true, err
	}
	return true, output(r)
}
func routePendingUpdate(dir string) (bool, error) {
	j, err := management.PendingUpdate(dir)
	if err != nil {
		return true, err
	}
	if j == nil {
		return false, nil
	}
	payload, _ := json.Marshal(management.UpdateRecoverPayload{Journal: *j})
	q := distribution.UpdateRequest{Protocol: j.Plan.Protocol, OperationID: j.Plan.OperationID, Action: distribution.UpdateRecover, ProductVersion: j.Plan.Package.Artifact.ProductVersion, Payload: payload}
	r, err := management.RouteUpdate(q, j.Plan)
	if err != nil {
		return true, err
	}
	return true, output(r)
}
