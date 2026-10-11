package taskqueue

import "testing"

func TestDefaults(t *testing.T) {
	t.Setenv("TEMPORAL_TASK_QUEUE", "")
	t.Setenv("TEMPORAL_WORKER_DEPLOYMENT_NAME", "")
	t.Setenv("TEMPORAL_DEPLOYMENT_NAME", "")
	t.Setenv("TEMPORAL_WORKER_BUILD_ID", "")

	if got := Name(); got != DemoTaskQueue {
		t.Fatalf("Name() = %q, want %q", got, DemoTaskQueue)
	}
	if got := DeploymentName(); got != DemoTaskQueue {
		t.Fatalf("DeploymentName() = %q, want %q", got, DemoTaskQueue)
	}
	if got := BuildID(); got != DemoBuildID {
		t.Fatalf("BuildID() = %q, want %q", got, DemoBuildID)
	}
	if got := ExecutionID("demo-ada-ab12"); got != "demo-ada-ab12-"+DemoTaskQueue {
		t.Fatalf("ExecutionID() = %q", got)
	}
}

func TestWorkspaceOverrides(t *testing.T) {
	t.Setenv("TEMPORAL_TASK_QUEUE", "serverless-webinar-abc1234567")
	t.Setenv("TEMPORAL_WORKER_DEPLOYMENT_NAME", "serverless-webinar-abc1234567")
	t.Setenv("TEMPORAL_DEPLOYMENT_NAME", "ignored-when-worker-name-is-set")
	t.Setenv("TEMPORAL_WORKER_BUILD_ID", "1.0.0")

	if got := Name(); got != "serverless-webinar-abc1234567" {
		t.Fatalf("Name() = %q", got)
	}
	if got := DeploymentName(); got != "serverless-webinar-abc1234567" {
		t.Fatalf("DeploymentName() = %q", got)
	}
	if got := BuildID(); got != "1.0.0" {
		t.Fatalf("BuildID() = %q", got)
	}
	if got := ExecutionID("demo-ada-ab12"); got != "demo-ada-ab12-serverless-webinar-abc1234567" {
		t.Fatalf("ExecutionID() = %q", got)
	}
	query := CountQuery("Running")
	want := `TaskQueue="serverless-webinar-abc1234567" AND ExecutionStatus="Running"`
	if query != want {
		t.Fatalf("CountQuery() = %q, want %q", query, want)
	}
}

func TestDeploymentNameFallback(t *testing.T) {
	t.Setenv("TEMPORAL_WORKER_DEPLOYMENT_NAME", "")
	t.Setenv("TEMPORAL_DEPLOYMENT_NAME", "serverless-webinar-abc1234567")
	if got := DeploymentName(); got != "serverless-webinar-abc1234567" {
		t.Fatalf("DeploymentName() = %q", got)
	}
}

func TestCountQueryEscapesQuotes(t *testing.T) {
	t.Setenv("TEMPORAL_TASK_QUEUE", `queue"x`)
	query := CountQuery("Running")
	want := `TaskQueue="queue\"x" AND ExecutionStatus="Running"`
	if query != want {
		t.Fatalf("CountQuery() = %q, want %q", query, want)
	}
}
