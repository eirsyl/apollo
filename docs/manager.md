# Manager

The manager is responsible for managing cluster operations by executing commands
in the cluster trough the agent nodes.

The manager consists of internal orchestration logic and two servers, one for grpc
communication and one for debugging, monitoring and API access.

The manager stores the internal state in memory and it's not possible to run the
manager in HA mode. This greatly complicates the implementation, the purpose of
this software is to manage redis clusters, not learn how to implement HA systems.