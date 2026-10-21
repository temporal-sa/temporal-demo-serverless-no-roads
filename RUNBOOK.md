# Runbook: Serverless Workers (No Roads) demo

This runbook is for presenting the demo. The [README](README.md) explains what it is and how it's deployed, and [METRICS.md](METRICS.md) explains where each number on the dashboard comes from.

## What it shows

The audience starts real Workflows from their phones, then you fire a burst of 30 or more at once. No Worker is running when the burst lands. Temporal Cloud sees the Task Queue back up, invokes the Worker as an AWS Lambda function, and the dashboard shows the backlog, the sync match rate, and the Lambda count move as the work drains.

The demo makes two problems visible:

- **Bursty work needs Workers now, and a fleet sized for the peak sits idle the rest of the time.** The burst in beat 3 arrives with no Worker fleet to absorb it. Temporal starts Lambda Workers for it, and nothing runs once it's done.
- **Something has to decide when to add Workers.** A CPU-based autoscaler can't see a Task Queue backlog. Here Temporal scales from the Task Queue itself: when a Task can't be handed straight to a waiting Worker, it invokes more Lambdas. Beats 3 and 4 show the sync match rate fall and recover.

## Before you start

- **On tmprl-demo.cloud** you need only a browser and a Temporal sign-in. The demo is at https://no-roads-serverless.tmprl-demo.cloud. The catalog runs the web app on Kubernetes and the Worker as a Lambda function in Worker Deployment `serverless-webinar`, Build ID `1.0.0`, with a 600-second function timeout. The catalog also links the slide deck, "Introduction to Serverless Workers".
- The shared demo uses one Task Queue, `serverless-webinar`, so its counts include anyone else running it at the same time. For a copy of your own, use the catalog card's isolated environment: it gets its own Task Queue, Worker Deployment and Lambda function, and the app reads the queue from `TEMPORAL_TASK_QUEUE`.
- **To run it locally** you need [Go](https://go.dev/dl/) 1.24 or later and the [Temporal CLI](https://docs.temporal.io/cli). There is no Lambda locally: a long-polling Worker with the same settings stands in for it, so the Lambda count stays at 1 and nothing scales. Use it to rehearse the screen and the talk track.
- Deploying your own Lambda and Worker Deployment takes an AWS account and a Temporal Cloud Namespace with an API key. The README's [Deploying the Lambda worker to AWS](README.md#deploying-the-lambda-worker-to-aws) section covers it.

## Start it

**On tmprl-demo.cloud:**

1. Open https://no-roads-serverless.tmprl-demo.cloud/?presenter=1 on the screen you'll share. The `?presenter=1` shows the **Presenter mode** panel. `Ctrl+Shift+P` also toggles it, unless your browser keeps that shortcut for itself.
2. Share the URL without `?presenter=1` with the audience (a QR code works well). Their view has no presenter panel.

A healthy start shows the **Submit a workflow** card, five stat cards, and an **Activity feed** that reads "Waiting for submissions...". **Sync match rate** reads "Waiting for first tasks…" when nothing has run in the last 30 seconds.

**Locally**, in three terminals from the repository root:

1. `temporal server start-dev`
2. `cd demo-app && go run ./localworker`
3. `cd demo-app && go run .`

The Worker logs `local worker started, polling task queue: serverless-webinar deployment: serverless-webinar build: 1.0.0`. Its first poll creates Worker Deployment `serverless-webinar`, but with no Current Version, so it gets no Tasks yet. Make its version current once per dev server:

```bash
temporal --disable-config-file worker deployment set-current-version \
  --deployment-name serverless-webinar --build-id 1.0.0 --yes
```

`--disable-config-file` keeps a Temporal CLI profile from sending the command to a Cloud Namespace. Open http://localhost:8080/?presenter=1. The Temporal UI is at http://localhost:8233.

## The walkthrough

The whole demo takes five to ten minutes, most of it audience participation. Each Workflow runs about 36 seconds, so the dashboard keeps moving for a minute or two after the last start.

### 1. The dashboard, before anyone submits

1. Show the five stat cards: **Running**, **Completed**, **Lambda** invocations, **Backlog** tasks queued, and **Sync match rate**.
2. Say: "There's no Worker fleet behind this page. The Worker is a Lambda function, and Temporal Cloud invokes it only when there's work on the Task Queue."
3. **Lambda** counts Workers that have polled in the last five minutes, so it can read above 0 for a few minutes after a previous run.

### 2. The audience starts Workflows

1. Ask the audience to open the link, type a name in **Your name or handle**, and press **Start workflow** (or `Enter`).
2. Each submission starts a `DemoWorkflow`. The page shows `✓ Workflow demo-<name>-<id>-serverless-webinar started for "<name>"`, the **Activity feed** adds a row, and **Running** climbs within a few seconds.
3. Say: "Every name is a real Workflow on Temporal Cloud. Each one runs four Activities, about 36 seconds in all, on whichever Lambda Worker Temporal started."
4. For proof, open the demo's Namespace in Temporal Cloud and search Workflows for an audience member's name. The Workflow ID starts with `demo-<name>`.

### 3. Fire a burst

1. In **Presenter mode**, set **Fire** to the number of Workflows (30 by default, 200 at most; larger values are capped at 200) and press **Flux the Capacitor!**
2. The panel shows "✓ Burst fired — 30 started", and the feed adds a "⚡ Seed burst" row. Within a few seconds **Backlog** and the **Task queue backlog** chart climb, and **Sync match rate** drops. Below 90% the card turns amber and reads "Pressure building — Temporal may invoke more workers"; below 60% it turns red and reads "Under load — Temporal is invoking Lambda to scale up".
3. **Lambda** and the **Lambda concurrency** chart rise as Temporal invokes more functions.
4. Say: "Each Lambda Worker takes five Activities at a time. Thirty Workflows don't fit, so Tasks wait in the queue. Temporal sees that a Task couldn't be handed straight to a Worker, and it invokes another Lambda."
5. For proof, show the Lambda function's **Monitor** tab in the AWS console, or list the system Workflow that does the scaling, if your CLI points at the demo's Namespace:

   ```bash
   temporal workflow list \
     --query 'TemporalNamespaceDivision = "TemporalWorkerControllerInstance"'
   ```

   Its ID is `temporal-sys-worker-controller-instance:serverless-webinar:1.0.0`.

### 4. The backlog drains

1. Keep talking for a minute. **Backlog** falls to 0, **Sync match rate** returns to green ("Workers keeping up"), **Running** falls and **Completed** climbs. Feed rows switch from running to completed.
2. Say: "Nobody configured a scaling policy, and there's no fleet to scale back down. When the queue is empty, the Lambdas stop and nothing is running."
3. **Lambda** falls later than the other numbers, for two reasons: each invocation keeps polling until shortly before the function's 600-second timeout, and a stopped Worker stays in the poller list for up to five minutes. Don't wait for it to reach 0 on stage.

### 5. Optional: each Workflow stays on the build that started it

1. Run `temporal workflow describe -w <workflow-id>` for one of the Workflows, with your CLI pointed at the demo's Namespace.
2. Under **Versioning Info** it shows Behavior `Pinned`, DeploymentName `serverless-webinar`, BuildId `1.0.0`.
3. Say: "Serverless Workers always use Worker Versioning. A new build is a new version: Workflows already running finish on the build they started on, and new ones start on the new build."

## Show it in code

All links point at commit [`0d1f338`](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/commit/0d1f338d33fc8a626753524e6cd682544e6573eb). Open a file locally with `code -g <path>:<line>`.

### Run a Worker on Lambda with one call

**When to use it:** after beat 3, or when someone asks what it takes to run a Worker on Lambda.

**Where it is:** [`lambda-worker/main.go` lines 48–86](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/lambda-worker/main.go#L48-L86) (`code -g lambda-worker/main.go:48`)

```go
version := worker.WorkerDeploymentVersion{
	DeploymentName: taskqueue.DeploymentName(),
	BuildID:        taskqueue.BuildID(),
}
// ...
lambdaworker.RunWorker(
	version,
	func(opts *lambdaworker.Options) error {
		opts.TaskQueue = taskqueue.Name()
		opts.WorkerOptions.MaxConcurrentActivityExecutionSize = cfg.MaxConcurrentActivityExecutionSize
		opts.WorkerOptions.MaxConcurrentWorkflowTaskExecutionSize = cfg.MaxConcurrentWorkflowTaskExecutionSize
		// ...
		opts.ClientOptions = workerconfig.BuildClientOptions()

		opts.RegisterWorkflow(workflows.DemoWorkflow)
		opts.RegisterActivity(&activities.Activities{})

		return nil
	},
)
```

**What you need to understand:**

- `lambdaworker.RunWorker` from the Go SDK's `contrib/aws/lambdaworker` package (v0.1.1 here) is the Lambda handler. On each invocation it connects to Temporal, starts a Worker, polls until shortly before the invocation deadline, then stops the Worker and closes the connection. Nothing carries over between invocations.
- With the defaults, it starts the shutdown 7 seconds before the function timeout (a 5-second Worker stop timeout plus 2 seconds for shutdown hooks). The catalog sets the timeout to 600 seconds, so each invocation polls for close to ten minutes even if the queue empties sooner.
- Registration is the same as a long-running Worker. The Workflow and Activities are the same code the local Worker in `demo-app/localworker` runs.
- The customer writes no scaling code and no invoker. Temporal Cloud invokes the function, through an IAM role in the customer's AWS account, when the Worker Deployment Version's Task Queue needs a Worker. That wiring is configuration on the Worker Deployment Version (the Lambda ARN and the role), not code.
- Each invocation gets its own Worker identity built from the Lambda request ID, so each one shows as a separate poller. That's what the **Lambda** stat counts.
- Serverless Workers are in Public Preview. Lambda caps an invocation at 15 minutes, and the package docs ask for a function timeout of at least the longest Activity `StartToCloseTimeout` plus the Worker stop timeout.

**Docs:** [Serverless Workers](https://docs.temporal.io/serverless-workers), [Serverless Workers on AWS Lambda](https://docs.temporal.io/serverless-workers/aws-lambda), [Go: Serverless Workers on AWS Lambda](https://docs.temporal.io/develop/go/workers/serverless-workers/aws-lambda)

### Pin each Workflow to the build that started it

**When to use it:** during beat 5, or when someone asks how a new deploy affects Workflows already running.

**Where it is:** [`lambda-worker/main.go` lines 67–73](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/lambda-worker/main.go#L67-L73) (`code -g lambda-worker/main.go:67`). The name and Build ID come from [`shared/taskqueue/taskqueue.go` lines 29–54](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/shared/taskqueue/taskqueue.go#L29-L54).

```go
// Opt the worker into deployment-based versioning. Without
// UseVersioning: true, Temporal won't route versioned tasks here.
opts.WorkerOptions.DeploymentOptions = worker.DeploymentOptions{
	UseVersioning:             true,
	Version:                   version,
	DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
}
```

**What you need to understand:**

- Serverless Workers require Worker Versioning. The compute provider (the Lambda ARN and the IAM role) is set on a Worker Deployment Version, which is how Temporal knows what to invoke. `RunWorker` turns versioning on by itself; the explicit `UseVersioning: true` here only repeats it.
- `DefaultVersioningBehavior: Pinned` applies to every Workflow type that doesn't set its own behavior. A Pinned Workflow completes on the version it started on, so a new build never replays old histories and the Workflow code needs no patching.
- New Workflows start on the Deployment's Current Version. A version that isn't Current gets no new Workflows, which is why the local setup runs `set-current-version` once. On tmprl-demo.cloud the operator sets it.
- The deployment name and Build ID come from environment variables, `serverless-webinar` and `1.0.0` when unset. The catalog's isolated environment suffixes the deployment name per workspace and keeps the Build ID.
- For production, the docs recommend pointing each Worker Deployment Version at a published Lambda version (a qualified ARN). With an unqualified ARN, a redeploy changes the code under Workflows already pinned to that version.

**Docs:** [Worker Versioning](https://docs.temporal.io/worker-versioning), [Serverless Workers and Worker Versioning](https://docs.temporal.io/serverless-workers/aws-lambda#worker-versioning)

### Keep each Lambda small so a burst needs more of them

**When to use it:** during beat 3, when someone asks why 30 Workflows are enough to make it scale.

**Where it is:** [`shared/workerconfig/workerconfig.go` lines 30–35](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/shared/workerconfig/workerconfig.go#L30-L35) (`code -g shared/workerconfig/workerconfig.go:30`), applied in [`lambda-worker/main.go` lines 59–60](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/lambda-worker/main.go#L59-L60)

```go
// Defaults — intentionally low to make backlog and sync match rate pressure
// visible during the demo. Set higher for production Lambda deployments.
const (
	DefaultMaxConcurrentActivity = 5
	DefaultMaxConcurrentWorkflow = 5
)
```

**What you need to understand:**

- Each invocation runs at most 5 Activities at a time. `WORKER_MAX_CONCURRENT_ACTIVITIES` and `WORKER_MAX_CONCURRENT_WORKFLOWS` override the defaults; the Worker logs the values it uses at startup.
- The low limit is a demo choice. It makes a burst of 30 overflow one Worker, so the backlog and the scaling are visible with a small audience.
- A customer sizes this to the function's memory and the Activity's work. `lambdaworker`'s own defaults are lower still: 2 Activities and 1 Activity poller per invocation.

### Send every Activity through the Task Queue

**When to use it:** when someone asks why there's a setting about eager Activities, or why the backlog would otherwise look empty.

**Where it is:** [`lambda-worker/main.go` lines 61–65](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/lambda-worker/main.go#L61-L65) (`code -g lambda-worker/main.go:61`), and the same setting in [`demo-app/localworker/main.go` line 36](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/demo-app/localworker/main.go#L36)

```go
// Disable eager activity dispatch so activity tasks flow through the
// task queue rather than being handed directly to the completing worker.
// Without this, TasksAddRate stays zero because tasks never touch the
// queue — making backlog and sync match rate metrics invisible.
opts.WorkerOptions.DisableEagerActivities = true
```

**What you need to understand:**

- With eager execution, the server can hand an Activity straight back to the Worker that completed the Workflow Task. Those Tasks skip the Task Queue, and the Task Queue's add and dispatch rates don't count them.
- Turning it off sends every Activity through the queue, where it can wait in the backlog and where any Worker can pick it up.
- `lambdaworker` already disables eager Activities for every Lambda Worker, so on Lambda this line repeats the default. On the local Worker it's needed.

**Docs:** [Task Queue statistics](https://docs.temporal.io/develop/worker-performance/task-queues#TasksAddRate-and-TasksDispatchRate)

### The Workflow doesn't know it runs on Lambda

**When to use it:** when someone asks what changes in their Workflow code to go serverless.

**Where it is:** [`shared/workflows/demo_workflow.go` lines 29–60](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/shared/workflows/demo_workflow.go#L29-L60) (`code -g shared/workflows/demo_workflow.go:29`)

```go
func DemoWorkflow(ctx workflow.Context, input DemoInput) (DemoOutput, error) {
	// ...
	ao := workflow.ActivityOptions{
		// Must exceed the sum of all three WorkDuration constants.
		StartToCloseTimeout: 60 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: fast processing activity
	var result string
	if err := workflow.ExecuteActivity(ctx, "ProcessSubmission", input).Get(ctx, &result); err != nil {
		return DemoOutput{}, err
	}

	// Steps 2-4: chained sleep activities — each holds a worker slot, driving
	// backlog depth and sync match rate pressure under load.
	var workResult string
	if err := workflow.ExecuteActivity(ctx, "SimulateWork1", input).Get(ctx, &workResult); err != nil {
		return DemoOutput{}, err
	}
	// ... SimulateWork2, SimulateWork3
	return DemoOutput{Message: result}, nil
}
```

**What you need to understand:**

- This is an ordinary Go Workflow. The same code runs on the local long-polling Worker and on Lambda; only the Worker's `main` differs.
- Each `SimulateWork` Activity sleeps for 12 seconds ([`shared/activities/demo_activities.go` lines 37–41](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/shared/activities/demo_activities.go#L37-L41)). Sleeping inside an Activity holds a Worker slot, which is what fills each Lambda. A `workflow.Sleep` would free the slot.
- The comment on `StartToCloseTimeout` is slightly off: the timeout applies to each Activity attempt separately, so 60 seconds only has to exceed the longest single Activity (12 seconds).
- The Activities use the default Retry Policy. If a Lambda stops mid-Activity, the attempt times out and Temporal retries it on another Worker.

### How the dashboard reads the Task Queue

**When to use it:** when someone asks whether the numbers on screen are real, or where to get them for their own system.

**Where it is:** [`demo-app/api/metrics.go` lines 171–224](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/demo-app/api/metrics.go#L171-L224) (`code -g demo-app/api/metrics.go:171`)

```go
for _, tqType := range tqTypes {
	resp, respErr := tc.WorkflowService().DescribeTaskQueue(ctx, &workflowservice.DescribeTaskQueueRequest{
		Namespace: namespace,
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: taskqueue.Name(),
		},
		TaskQueueType: tqType,
		ReportStats:   true,
	})
	// ...
		totalBacklog += resp.Stats.ApproximateBacklogCount
		totalTasksAddRate += resp.Stats.TasksAddRate
		totalTasksDispatchRate += resp.Stats.TasksDispatchRate
	// ...
	if tqType == enumspb.TASK_QUEUE_TYPE_ACTIVITY {
		pollerCount = len(resp.Pollers)
	}
}
// ...
totalBacklogIncRate := totalTasksAddRate - totalTasksDispatchRate
rate := (1.0 - float64(totalBacklogIncRate)/float64(totalTasksAddRate)) * 100.0
```

**What you need to understand:**

- Every number comes from Temporal. **Backlog** and **Sync match rate** come from `DescribeTaskQueue` with `ReportStats`, once for Workflow Tasks and once for Activity Tasks. **Running** and **Completed** are Visibility counts filtered to the Task Queue. The app caches the result for 2 seconds and the page polls every 2 seconds.
- **Sync match rate** is the app's own figure: dispatch rate divided by add rate over the last 30 seconds, clamped to 0–100%. It tracks whether the backlog is growing. It isn't the signal Temporal's scaler reads; the scaler gets sync match failures from the Matching Service directly.
- **Lambda** is the number of Activity pollers. Each invocation polls with its own identity, so it approximates live invocations, but the server keeps a poller in the list until five minutes after its last poll. On the way down it lags.
- These are the same APIs a customer can use for their own dashboards or autoscaling. Nothing here drives the scaling; Temporal Cloud does that without this app.

**Docs:** [Task Queue performance](https://docs.temporal.io/develop/worker-performance/task-queues), [`temporal task-queue describe`](https://docs.temporal.io/cli/command-reference/task-queue#describe)

### What's demo-only

When someone asks what they'd change for their own system, these are the parts to leave out:

- **The Activities sleep.** [`shared/activities/demo_activities.go` lines 56–66](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/shared/activities/demo_activities.go#L56-L66): each one holds a slot for 12 seconds and returns a string. Real Activities replace them; the Workflow doesn't change.
- **The low concurrency limits** exist so a small burst overflows one Worker. See [Keep each Lambda small](#keep-each-lambda-small-so-a-burst-needs-more-of-them).
- **The presenter burst.** [`demo-app/api/seed.go` lines 67–92](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/demo-app/api/seed.go#L67-L92) starts up to 200 Workflows in parallel goroutines. It has no authentication; anyone who finds `/api/seed` can call it.
- **The dashboard and its metrics endpoint.** Temporal Cloud scales the Lambdas without them.
- **The Task Queue suffix in Workflow IDs.** [`shared/taskqueue/taskqueue.go` lines 56–60](https://github.com/temporal-sa/temporal-demo-serverless-no-roads/blob/0d1f338d33fc8a626753524e6cd682544e6573eb/shared/taskqueue/taskqueue.go#L56-L60) appends the queue name so isolated workspaces sharing one Namespace never collide.

## Reset and tear down

- **Between runs:** there is no reset button. Running Workflows finish on their own about 36 seconds after their last Task starts. To clear them at once, terminate them with one batch operation:

  ```bash
  temporal workflow terminate \
    --query 'TaskQueue="serverless-webinar" AND ExecutionStatus="Running"' \
    --reason "reset demo" --yes
  ```

  On the shared demo this also terminates anyone else's running Workflows. **Completed** counts every completed Workflow on the queue that the Namespace still retains, so it doesn't go back to 0.
- **Locally:** stop the app, the Worker and the dev server with `Ctrl-C`. The dev server keeps nothing, so the next start begins at 0 (and needs `set-current-version` again).
- **On tmprl-demo.cloud:** the shared demo is always on; leave it running. Nothing in the demo starts Workflows on its own. Each Lambda invocation stops on its own before its 600-second timeout, and Temporal invokes no more once the queue is empty.

## When something goes wrong

| What you see | Likely cause | Fix |
|---|---|---|
| Locally, **Running** and **Backlog** climb but nothing completes | The Worker Deployment has no Current Version, so the versioned Worker gets no Tasks | Run the `set-current-version` command in [Start it](#start-it) |
| The **Presenter mode** panel isn't there | The URL has no `?presenter=1` | Add `?presenter=1`, or press `Ctrl+Shift+P` |
| "Network error — is the server running?" | The demo app is down | Locally, start `go run .` in `demo-app`; on tmprl-demo.cloud, check the project's status in the catalog |
| All stat cards show 0 and nothing moves after a submit | The app can't reach Temporal | Check the app's log for `metrics:` errors and its `TEMPORAL_*` environment |
| **Sync match rate** shows "—" and "Waiting for first tasks…" | No Tasks were added in the last 30 seconds | Expected when idle; submit a name or fire a burst |
| **Backlog** stays up after you terminate Workflows | The queue still holds Tasks for the terminated Workflows until a Worker polls and drops them | Fire a small burst or wait; the next Worker drains them |
| **Lambda** stays above 0 long after the burst | Invocations poll until near their timeout, and pollers stay listed for five minutes | Expected; don't wait for it on stage |
| Fewer Workflows start than you typed in **Fire** | The burst is capped at 200 | Fire again |

## Keeping this runbook current

The code links point at a fixed commit. When a linked file changes, update its link, line numbers, and snippet together. Before an event, list the linked files that have changed:

```sh
grep -oE 'blob/[0-9a-f]{40}/[^#)]+' RUNBOOK.md | sort -u | while IFS=/ read -r _ sha file; do
  git diff --quiet "$sha" HEAD -- "$file" || echo "changed since ${sha:0:7}: $file"
done
```
