# taskcluster-cli
Taskcluster CLI is a command line interface for creating and managing 
tasks submitted to taskcluster.

# Requirements

 - node 0.10.30 or greater

# Installation
1. Run `npm install -g taskcluster-cli` to install taskcluster-cli and required dependencies.

# Configuration

Add taskcluster credential environment variables.  This is best done in your shell profile.

```sh
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
taskcluster run --owner=your@email.com --provisioner-id=aws-provisioner --worker-type=cli ubuntu:14.04 ls

[taskcluster] taskId: 82LOBaruRZaGSXqXc3U6rA, workerId: i-56443d59

ubuntu:14.04 exists in the cache.
bin   dev  home  lib64	mnt  proc  run	 srv  tmp  var
boot  etc  lib	 media	opt  root  sbin  sys  usr
[taskcluster] Successful task run with exit code: 0 completed in 1.337 seconds
```
