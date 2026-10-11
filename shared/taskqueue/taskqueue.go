package taskqueue

import (
	"os"
	"strings"
)

const (
	// DemoTaskQueue is the shared demo's task queue and Worker Deployment name.
	// An isolated workspace overrides the queue and the deployment name.
	DemoTaskQueue = "serverless-webinar"
	// DemoBuildID is the shared demo's Worker Deployment build ID.
	// Workspaces keep this build ID; only the deployment name is suffixed.
	DemoBuildID = "1.0.0"
)

// Name is the task queue this process polls or starts workflows on.
// The registry injects TEMPORAL_TASK_QUEUE into an isolated workspace.
// When it is unset, this is the shared demo queue.
//
// Do not call this from workflow code. The queue is process configuration.
func Name() string {
	if v := os.Getenv("TEMPORAL_TASK_QUEUE"); v != "" {
		return v
	}
	return DemoTaskQueue
}

// DeploymentName is the Worker Deployment this process claims.
// The operator sets TEMPORAL_WORKER_DEPLOYMENT_NAME and TEMPORAL_DEPLOYMENT_NAME
// to the same value. An isolated workspace suffixes the shared name.
//
// Do not call this from workflow code.
func DeploymentName() string {
	if v := os.Getenv("TEMPORAL_WORKER_DEPLOYMENT_NAME"); v != "" {
		return v
	}
	if v := os.Getenv("TEMPORAL_DEPLOYMENT_NAME"); v != "" {
		return v
	}
	return DemoTaskQueue
}

// BuildID is the Worker Deployment build ID this process claims.
// The operator sets TEMPORAL_WORKER_BUILD_ID. Build IDs are not suffixed
// per workspace.
//
// Do not call this from workflow code.
func BuildID() string {
	if v := os.Getenv("TEMPORAL_WORKER_BUILD_ID"); v != "" {
		return v
	}
	return DemoBuildID
}

// ExecutionID includes the active task queue so workflow IDs stay distinct
// when isolated workspaces share the demo's Temporal namespace.
func ExecutionID(suffix string) string {
	return suffix + "-" + Name()
}

// CountQuery is a visibility count filter for this process's task queue.
func CountQuery(status string) string {
	return `TaskQueue="` + escapeVisibility(Name()) + `" AND ExecutionStatus="` + escapeVisibility(status) + `"`
}

func escapeVisibility(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
}
