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
. ~/tmp/kind/kubeconfigs/awsvm/awsvm_clustermaps.sh
```
Now we can use our laptop to access the clusters in VM
```
kubectl get pods --kubeconfig $cluster1 -A
kubectl get pods --kubeconfig $cluster2 -A
kubectl get pods --kubeconfig $cluster3 -A
```

**Run vl3 test on you laptop**
go to the nsm-nse local repo
```
cd ~/go/src/github.com/cisco-app-networking/nsm-nse
```
run the script
```
NSE_HUB=ciscoappnetworking NSE_TAG=master KUBECONFDIR=~/tmp/kind/kubeconfigs/awsvm/kubeconfigs/nsm build/ci/runner/run_vl3.sh
```
clean up
```
kubectl delete deployment helloworld-ucnf --kubeconfig $cluster2
kubectl delete deployment vl3-nse-ucnf --kubeconfig $cluster2
kubectl delete namespace nsm-system --kubeconfig $cluster2
kubectl delete namespace spire --kubeconfig $cluster2
kubectl delete deployment helloworld-ucnf --kubeconfig $cluster3
kubectl delete deployment vl3-nse-ucnf --kubeconfig $cluster3
kubectl delete namespace nsm-system --kubeconfig $cluster3
kubectl delete namespace spire --kubeconfig $cluster3
```
after clean up, if we want to test, we will need to build the image again
first login
```
docker login --username=shanchunyang0919
```
make&push the image to the dockerhub, and then we can run the **run_vl3.sh** script again.
```
ORG=shanchunyang0919 TAG=master make docker-vl3
docker image push shanchunyang0919/vl3_ucnf-nse:master
```
run the script with the env variables tiding to our dockerhub
```
NSE_HUB=shanchunyang0919 NSE_TAG=master KUBECONFDIR=~/tmp/kind/kubeconfigs/awsvm/kubeconfigs/nsm build/ci/runner/run_vl3.sh
```
If we want to test anything, just repeat the previous steps

**Check if the service account is updated**
```
kubectl describe sa/ucnf-service-account --kubeconfig $cluster2
kubectl describe sa/ucnf-service-account --kubeconfig $cluster3
```
