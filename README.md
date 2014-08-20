# taskcluster-cli
that you'd normally have to do by hand. There are tasks that are common
Taskcluster CLI is a command line interface for creating and managing 
tasks submitted to taskcluster.

# Requirements

 - node 0.11x or greater

# Installation
1. Clone a copy of this repo
2. Run `npm install <path to repo>` to install taskcluster-cli and required dependencies.
3. If not installing globally, add the node_modules/.bin path to your shell.

# Configuration

Add taskcluster credential environment variables.  This is best done in your shell profile.

```sh
export TASKCLUSTER_TASK_OWNER=...
export TASKCLUSTER_CLIENT_ID=...
export TASKCLUSTER_ACCESS_TOKEN=...
```

# Usage

```sh
taskcluster run --provisioner-id=<instance provisioner> --worker-type=<worker type> <image> <commands>
```

# Example

Create a task that will count the number of files in a directory on Ubuntu 14.04.

```sh
taskcluster run --provisioner-id=aws-provisioner --worker-type=cli ubuntu:14.04 'find /bin -type f -print | wc -l'

Task Created.
Task ID: xquu2goHS3-pexVC9w4dmw
Task State: Pending
Task Completed
[taskcluster] taskId: xquu2goHS3-pexVC9w4dmw, workerId: i-efa604e0 

ubuntu:14.04 exists in the cache.
find /bin -type f -print | wc -l
+ find /bin -type f -print
+ wc -l
100
[taskcluster] Successful task run with exit code: 0 completed in 0.984 seconds
```
