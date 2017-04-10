### Deployment
Included in this directory are some template files for running Burrow in a 
cluster orchestration environment. [start_in_cluster](start_in_cluster) 
is a general purpose script and the files in [kubernetes](kubernetes) are some
example Service and Deployment files that illustrates its possible usage.

#### start_in_cluster
[start_in_cluster](start_in_cluster) takes its parameters as environment variables.

You can find the variables used at the top of the file along with their defaults.

#### Kubernetes
[all_nodes.yml](kubernetes/all_nodes.yaml) is a Kubernetes Service definition
that launches an entire network of nodes based on Deployment definitions like the
example [node000-deploy.yaml](kubernetes/node000-deploy.yaml). Each validating
node should have it's own Deployment defintion like the one found in 
[node000-deploy.yaml](kubernetes/node000-deploy.yaml)

