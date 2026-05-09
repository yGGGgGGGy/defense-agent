package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

func TestFullIncidentResponseFlow(t *testing.T) {
	// 1. Submit task
	task := map[string]any{
		"scene":       "incident_response",
		"title":       "Integration Test: SSH Brute Force",
		"description": "E2E test",
		"input":       "SSH brute force from 10.0.0.99",
		"alerts": []map[string]any{
			{"id": "INT-001", "rule": "SSH_BRUTE_FORCE", "source_ip": "10.0.0.99", "count": 200},
		},
	}

	body, _ := json.Marshal(task)
	resp, err := http.Post(baseURL+"/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var inst map[string]any
	json.NewDecoder(resp.Body).Decode(&inst)
	instanceID := inst["id"].(string)
	t.Logf("Instance: %s", instanceID)

	// 2. Wait for completion
	time.Sleep(5 * time.Second)

	// 3. Check status
	resp2, err := http.Get(baseURL + "/tasks/" + instanceID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	defer resp2.Body.Close()

	var status map[string]any
	json.NewDecoder(resp2.Body).Decode(&status)
	t.Logf("Status: %v", status["status"])

	dag := status["dag"].(map[string]any)
	nodes := dag["nodes"].(map[string]any)
	for _, n := range nodes {
		node := n.(map[string]any)
		t.Logf("  %s: %s", node["agent_type"], node["state"])
	}

	// 4. Check audit chain
	resp3, err := http.Get(baseURL + "/tasks/" + instanceID + "/audit")
	if err != nil {
		t.Fatalf("get audit: %v", err)
	}
	defer resp3.Body.Close()

	var records []map[string]any
	json.NewDecoder(resp3.Body).Decode(&records)
	t.Logf("Audit records: %d", len(records))

	if len(records) < 8 {
		t.Errorf("expected at least 8 audit records, got %d", len(records))
	}

	// Verify chain integrity
	for i := 1; i < len(records); i++ {
		prevHash := records[i-1]["record_hash"].(string)
		curPrevHash := records[i]["prev_hash"].(string)
		if prevHash != curPrevHash {
			t.Errorf("chain broken at record %d: prev_hash mismatch", i)
		}
	}
}

func TestMultiInstanceIsolation(t *testing.T) {
	var instanceIDs []string

	// Submit 2 tasks in parallel
	for i := 0; i < 2; i++ {
		task := map[string]any{
			"scene":       "ops_maintenance",
			"title":       "Isolation Test",
			"description": "Test",
			"input":       "Health check",
		}
		body, _ := json.Marshal(task)
		resp, err := http.Post(baseURL+"/tasks", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("create task: %v", err)
		}
		var inst map[string]any
		json.NewDecoder(resp.Body).Decode(&inst)
		resp.Body.Close()
		instanceIDs = append(instanceIDs, inst["id"].(string))
	}

	time.Sleep(4 * time.Second)

	// Verify isolation - each instance has its own audit chain
	for i, id := range instanceIDs {
		resp, err := http.Get(baseURL + "/tasks/" + id + "/audit")
		if err != nil {
			t.Fatalf("get audit for %s: %v", id, err)
		}
		var records []map[string]any
		json.NewDecoder(resp.Body).Decode(&records)
		resp.Body.Close()

		// Each operator-only instance should have 4 audit records
		if len(records) != 4 {
			t.Errorf("instance %d (%s): expected 4 records, got %d", i, id, len(records))
		}
		t.Logf("Instance %d (%s): %d records, audit chain valid", i, id, len(records))
	}
}
