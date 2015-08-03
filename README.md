# taskcluster-cli
TaskCluster CLI is a command line interface for creating and managing
tasks submitted to TaskCluster.

# Requirements

 - node 0.10.30 or greater

# Installation
1. Run `npm install -g taskcluster-cli` to install taskcluster-cli and required dependencies.

# Configuration

Add TaskCluster credential environment variables.  This is best done in your shell profile.

```sh
export TASKCLUSTER_CLIENT_ID=...
export TASKCLUSTER_ACCESS_TOKEN=...
export TASKCLUSTER_CERTIFICATE=... # the full JSON string
```

Or you can run `taskcluster login` which will authenticate you with temporary
credentials that will then be stored in a configuration. This is the preferred
way of authentication.

# Usage

```sh
taskcluster run --provisioner-id=<instance provisioner> --worker-type=<worker type> <image> <command>
```

# Example

Create a task that will list the files in a directory on Ubuntu 14.04.

```sh
taskcluster run --owner=your@email.com --provisioner-id=aws-provisioner --worker-type=cli ubuntu:14.04 ls
Task Created.
Task ID: 82LOBaruRZaGSXqXc3U6rA
Task State: Pending
[taskcluster] taskId: 82LOBaruRZaGSXqXc3U6rA, workerId: i-56443d59

ubuntu:14.04 exists in the cache.
bin   dev  home  lib64	mnt  proc  run	 srv  tmp  var
boot  etc  lib	 media	opt  root  sbin  sys  usr
[taskcluster] Successful task run with exit code: 0 completed in 1.337 seconds
```

Create a task that will list the packages installed on Ubuntu 14.04
```sh
taskcluster run --owner=your@email.com --provisioner-id=aws-provisioner --worker-type=cli ubuntu:14.04 -- /usr/bin/dpkg --list

Task Created.
Task ID: 5y3HM2dcRSWbegZp0MS-hg
Task State: Pending
[taskcluster] taskId: 5y3HM2dcRSWbegZp0MS-hg, workerId: i-aa67c563

Desired=Unknown/Install/Remove/Purge/Hold
| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend
|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)
||/ Name           Version      Architecture Description
+++-==============-============-============-=================================
ii  adduser        3.113+nmu3ub all          add and remove users and groups
....
[taskcluster] Successful task run with exit code: 0 completed in 10.279 seconds
Task Completed
```
