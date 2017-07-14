# k8s-monitor
Simple Golang client for watching Kubernetes Events. Can be run out of cluster or in cluster.

Currently I'm using this in a cluster that has a Fluentd -> Loggly daemon set running. This allows me to log Kubernetes events which Fluentd picks up and sends to Loggly. 

# Out of cluster
Looks for ~/.kube/config and uses the context set to connect to the cluster

# In cluster
If a kubectl config file cannot be found it will create an in cluster connection using the namespace service account

# Permissions
If Kubernetes authorization is enabled the account being used will need to have access to listen for events on all namespaces.