# vL3 NSE
This folder contains a Helm chart for the vL3 NSE. This document describes the most
important parts that may be needed in order to tweak the NSE deployment.

# Approach

**Trace where this helm chart is used.**

Etherlab Reference: https://etherpad.ci.ciscolabs.com/p/ETI_lab

First, ssh intoec2 instance
```
sudo ssh -I my_vm33.pem ubuntu@ec2-18-206-74-56.compute-1.amazonaws.com
```
After installing all the prequesites, goes to the wcm topo script foler
```
cd ~/go/src/github.com/cisco-app-networking/wcm-system/system_topo
```
Brings up the cluster (or delete)
```
sudo ./setup_kind_clusters.sh --kind-hostip=172.31
sudo ./setup_kind_clusters.sh --kind-hostip=172.31 --delete
```
**On your laptop**
```
mkdir -p ~/tmp/kind/kubeconfigs/awsvm)
```
copy kubeconfigs from AWS vm to the directory
```
sudo scp -r -i ~/.ssh/my_vm33.pem ubuntu@ec2-18-206-74-56.compute-1.amazonaws.com:kubeconfigs ~/tmp/kind/kubeconfigs/awsvm
```
Edit /etc/hosts to add lines
```
34.237.51.204 kind-1
34.237.51.204 kind-2
34.237.51.204 kind-3
```
Edit each of the SCP'd kubeconfigs to change the server line to use the hostname kind-1, 2, or 3, e.g.:
```
server: https://kind-1:38791
```
and kind-2.kubeconfig:
```
server: https://kind-2:38792
```
Create a "clustermaps" file that looks like this-- ~/tmp/kind/kubeconfigs/awsvm/lab_clustermaps.sh
```
cluster1=/Users/<YOUR_USERNAME>/tmp/kind/kubeconfigs/awsvm/kubeconfigs/central/kind-1.kubeconfig
cluster2=/Users/<YOUR_USERNAME>/tmp/kind/kubeconfigs/awsvm/kubeconfigs/nsm/kind-2.kubeconfig
cluster3=/Users/<YOUR_USERNAME>/tmp/kind/kubeconfigs/awsvm/kubeconfigs/nsm/kind-3.kubeconfig
```
Souce
```
 .~/tmp/kind/kubeconfigs/awsvm/lab_clustermaps.sh
```
Now we can use our laptop to access the clusters in VM
```
kubectl get pods --kubeconfig kind-1.kubeconfig
```
