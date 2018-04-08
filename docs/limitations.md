# Limitations

* Cluster creation based on amount of online empty nodes
* Not able to handle the case where two different clusters is reporting to the same manager
* Simple slot allocation algorithm
* Cluster check stops if nodes is offline

# Further work

* Downscaling?
* Data rebalancing?
* Automatic scaling in elastic environments (aws/kubernetes)
* Smarter resource allocation scheduling
* Manager and agent in HA mode
* Real world testing
* More advanced management interface