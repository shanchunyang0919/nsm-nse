#!/usr/bin/env bash
set -euo pipefail

MANIFEST=${MANIFEST:-kind-1-demo}
MANIFESTDIR=${TESTDIR:-${GOPATH}/src/github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_env_setup}
KCONFIG=${KCONFIG:-cluster1-demo}

echo
echo "CLEAN UP"
echo "---------------------"

kind delete cluster --name ${MANIFEST}
rm ${MANIFESTDIR}/${KCONFIG}

