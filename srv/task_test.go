package srv

import (
	"testing"
)

func TestTaskCRUD(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Task Test", "client_slug": "task", "project_slug": "test", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create task
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/tasks", map[string]any{
		"title":       "First Task",
		"assignee":    "Alice",
		"sort_order":  1,
		"status":      "pending",
		"curr_due":    "2025-02-01",
		"orig_budget": 500.0,
		"curr_budget": 500.0,
	})
	if resp.StatusCode != 201 {
		t.Fatalf("create task: expected 201, got %d", resp.StatusCode)
	}
	var task map[string]any
	decodeJSON(t, resp, &task)
	taskID := itoa(int64(task["ID"].(float64)))

	if task["Title"] != "First Task" {
		t.Errorf("expected title 'First Task', got %v", task["Title"])
	}
	if task["Assignee"] != "Alice" {
		t.Errorf("expected assignee 'Alice', got %v", task["Assignee"])
	}

	// List tasks
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("list tasks: expected 200, got %d", resp.StatusCode)
	}
	var tasks []map[string]any
	decodeJSON(t, resp, &tasks)
	if len(tasks) != 32 {
		t.Errorf("expected 32 tasks including seeded workflow, got %d", len(tasks))
	}
	foundCreated := false
	for _, got := range tasks {
		if got["ID"] == task["ID"] {
			foundCreated = true
			break
		}
	}
	if !foundCreated {
		t.Fatalf("expected created task id %v in task list", task["ID"])
	}

	// Update task
	resp = apiRequestAdmin(t, ts, "PUT", "/api/tasks/"+taskID, map[string]any{
		"title":         "Updated Task",
		"assignee":      "Bob",
		"status":        "active",
		"sort_order":    99,
		"actual_budget": 250.0,
	})
	if resp.StatusCode != 200 {
		t.Fatalf("update task: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify update
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	decodeJSON(t, resp, &tasks)
	updatedFound := false
	for _, got := range tasks {
		if got["ID"] == task["ID"] {
			updatedFound = true
			if got["Title"] != "Updated Task" {
				t.Errorf("expected updated title, got %v", got["Title"])
			}
			if got["Status"] != "active" {
				t.Errorf("expected status 'active', got %v", got["Status"])
			}
		}
	}
	if !updatedFound {
		t.Fatalf("expected updated task id %v in task list", task["ID"])
	}

	// Delete task
	resp = apiRequestAdmin(t, ts, "DELETE", "/api/tasks/"+taskID, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("delete task: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify deletion
	resp = apiRequestAdmin(t, ts, "GET", "/api/projects/"+pid+"/tasks", nil)
	decodeJSON(t, resp, &tasks)
	if len(tasks) != 31 {
		t.Errorf("expected 31 seeded tasks after delete, got %d", len(tasks))
	}
	for _, got := range tasks {
		if got["ID"] == task["ID"] {
			t.Fatalf("expected deleted task id %v to be absent", task["ID"])
		}
	}
}

func TestTaskMilestone(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	// Create project
	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Milestone Test", "client_slug": "m", "project_slug": "t", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create milestone task
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/tasks", map[string]any{
		"title":       "Big Milestone",
		"is_milestone": 1,
		"sort_order":  1,
	})
	var task map[string]any
	decodeJSON(t, resp, &task)

	if task["IsMilestone"].(float64) != 1 {
		t.Errorf("expected is_milestone=1, got %v", task["IsMilestone"])
	}
}

func TestTaskDefaultStatus(t *testing.T) {
	_, ts, cleanup := testServer(t)
	defer cleanup()

	resp := apiRequestAdmin(t, ts, "POST", "/api/projects", map[string]string{
		"name": "Status Test", "client_slug": "s", "project_slug": "t", "start_date": "2025-01-01",
	})
	var project map[string]any
	decodeJSON(t, resp, &project)
	pid := itoa(int64(project["ID"].(float64)))

	// Create task without status - should default to "pending"
	resp = apiRequestAdmin(t, ts, "POST", "/api/projects/"+pid+"/tasks", map[string]any{
		"title": "No Status",
	})
	var task map[string]any
	decodeJSON(t, resp, &task)

	if task["Status"] != "pending" {
		t.Errorf("expected default status 'pending', got %v", task["Status"])
	}
}
