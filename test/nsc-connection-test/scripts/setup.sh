#!/usr/bin/env bash
set -euo pipefail

NSE=${NSE:-1}
MANIFEST=${MANIFEST:-kind-1-demo}
KCONFIG=${KCONFIG:-cluster1-demo}
IPAM=${IPAM:-172.100.0.0/16}
TESTDIR=${TESTDIR:-${GOPATH}/src/github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test}

echo ${IPAM} ${MANIFEST}
echo "NSEs: ${NSE}"

cd ${TESTDIR}/test_env_setup

kind create cluster --name ${MANIFEST} --config ${MANIFEST}.yaml

echo
echo "LIST CLUSTERS"
echo "---------------------"
kind get clusters

kind get kubeconfig --name=${MANIFEST} > ${KCONFIG}

KUBEPATH=$GOPATH/src/github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_env_setup/${KCONFIG}

echo
echo "INSTALL NSM"
echo "---------------------"
cd ~/go/src/github.com/cisco-app-networking/nsm-nse
KCONF=${KUBEPATH} scripts/vl3/nsm_install_interdomain.sh
kubectl get pods -A

echo
echo "INSTALL VL3-NSE"
echo "---------------------"

REMOTE_IP=${IPAM} KCONF=${KUBEPATH} PULLPOLICY=Always NSEREPLICAS=${NSE} scripts/vl3/vl3_interdomain.sh

cd ${TESTDIR}

# Enviroment varibles for go test
INIT=${INIT:-on}
TIMEOUT=${TIMEOUT:-300}
NSE_LOG=${NSE_LOG:-30}
NSMGR_LOG=${NSMGR_LOG:-30}
PING_LOG=${PING_LOG:-on}
LOG=${LOG:-on}

echo
echo "NSC CONNECTION TEST START"
echo "---------------------"

INIT=${INIT} TIMEOUT=${TIMEOUT} LOG=${LOG} NSMGR_LOG=${NSMGR_LOG} PING_LOG=${PING_LOG} NSE_LOG=${NSE_LOG:-on} go test

