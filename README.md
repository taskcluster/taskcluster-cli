# taskcluster-cli

## Requirements

 - node 0.11 or greater

## Usage

First you need to export your taskcluster credentials:

```sh
export TASKCLUSTER_CLIENT_ID=...
export TASKCLUSTER_ACCESS_TOKEN=...
```

This is best done in you zsh/bash/sh profile...

```sh
# from the root of the project
node --harmony bin/taskcluster-run.js <docker_image> <args....>
```
