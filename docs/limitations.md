# Limitations

* Cluster creation based on amount of online empty nodes
* Not able to handle the case where two different clusters is reporting to the same manager
* Simple slot allocation algorithm
* Cluster check stops if nodes is offline
* Cannot fix slot coverage if keys exists on multiple nodes
* Does not watch executing tasks
* Not verifying results of executions
