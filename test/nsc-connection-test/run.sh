#!/usr/bin/env bash

set -euo pipefail

IPAM=${1:-172.100.0.0/16}
MANIFEST=${2:-kind-1-demo}
KCONFIG=${3:-cluster1-demo}
TESTDIR=${GOPATH}/src/github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_env_setup/


cd ${TESTDIR}

kind create cluster --name ${MANIFEST} --config ${MANIFEST}.yaml

echo
echo "LIST CLUSTERS"
echo "---------------------"
kind get clusters

kind get kubeconfig --name=${MANIFEST} >${KCONFIG}

KUBEPATH=$GOPATH/src/github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_env_setup/${KCONFIG}

echo
echo "INSTALL NSM"
echo "---------------------"
cd ~/go/src/github.com/cisco-app-networking/nsm-nse
KCONF=${KUBEPATH} scripts/vl3/nsm_install_interdomain.sh
kubectl get pods -A

REMOTE_IP=${IPAM} KCONF=${KUBEPATH} PULLPOLICY=Always NSEREPLICAS=2 scripts/vl3/vl3_interdomain.sh

cd ${TESTDIR}

BUSYBOX_SVC_PATH=./test_env_setup/vl3-busybox-svc.yaml

kubectl apply -f ${BUSYBOX_SVC_PATH}

go run main.go -apply -re=10




#echo
#echo "CLEAN UP"
#echo "---------------------"
#kind delete cluster --name ${MANIFEST}
