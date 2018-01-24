# Apollo [![CircleCI](https://circleci.com/gh/eirsyl/apollo.svg?style=svg&circle-token=112f280e9b22239b2ee800ca1f4f1705ed29ddf2)](https://circleci.com/gh/eirsyl/apollo)
> Automatic cluster manager for Redis Cluster

## Architecture

The cluster manager consists of multiple services. 

* manager - The primary coordinator service responsible for orchestrating cluster operations
* agent - Metric collector and coordinator for each Redis instance

The agent service runs as a sidecar deployed together with each Redis instance.
Only one manager should run at the time.

## CLI

```
./apollo --help

Apollo is a Redis Cluster manager that aims to lighten the operational burden
on cluster operators. The cluster manager watches the Redis cluster for possible
issues or reduced performance and tries to fix these in the best possible way.

Usage:
  apollo [command]

Available Commands:
  agent       Start the instance agent functionality
  chaos       Destroy Redis instances to test the cluster manager
  help        Help about any command
  manager     Start the cluster manager functionality

Flags:
  -d, --debug     enable debug mode
  -h, --help      help for apollo
      --version   version for apollo

Use "apollo [command] --help" for more information about a command.
```

## Getting started

```
git clone git@github.com:eirsyl/apollo.git
make
```
