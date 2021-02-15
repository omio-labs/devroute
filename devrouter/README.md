# Motivation

Envoy can route traffic to other services within the mesh and it requires no setup. However, when sending traffic out of the mesh, we need to specify a cluster (see [Cluster Manager](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/cluster_manager)) and we have to do it in advance, that is, before 
the user sends the request, we need to know what's the laptops `ip:port` and create a cluster with those parameters. This defeats the purpose of automation and low-entry barrier for this contract to be effective, i.e.: whenever you need to test something, you have to take steps to configure infrastructure. 

For these reasons we decided to create this simple proxy in QA whose only responsibility is to parse the 
contract and proxy the request to the corresponding `host:ip`. This proxy enforce validations that keep the invariants of this contract. 

