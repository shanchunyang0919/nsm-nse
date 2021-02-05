#!/usr/bin/env bash

set -euo pipefail

cd $GOPATH/src/github.com/cisco-app-networking/wcm-system/system_topo

svcname=${1:-bar}
ipamprefix=${2:-172.100.0.0/16}

echo
echo "SETUP SYSTEM TOPO"
echo "-----------------"
./setup_system_topo.sh

echo
echo "CREATE CONNECT DOMAIN $svcname"
echo "---------------------"
./create_connectdomain.sh --name="$svcname" --ipam-prefix="$ipamprefix"

echo
echo "DEPLOY DEMO APP $svcname"
echo "------------------------"
./deploy_demo_app.sh --service-name="$svcname" --nsc-delay=5