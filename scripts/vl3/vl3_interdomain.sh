#!/bin/bash
COMMAND_ROOT=$(dirname "${BASH_SOURCE}")
print_usage() {
  echo "$(basename "$0")
Usage: $(basename "$0") [options...]
Options:
  --nse-hub=STRING          Hub for vL3 NSE images
                            (default=\"ciscoappnetworking\", environment variable: NSE_HUB)
  --nse-tag=STRING          Tag for vL3 NSE images
                            (default=\"master\", environment variable: NSE_TAG)
" >&2
}

NSE_HUB=${NSE_HUB:-"ciscoappnetworking"}
NSE_TAG=${NSE_TAG:-"master"}
PULLPOLICY=${PULLPOLICY:-IfNotPresent}
INSTALL_OP=${INSTALL_OP:-apply}
SERVICENAME=${SERVICENAME:-vl3-service}
NAMESPACE=${NAMESPACE:-"wcm-system"}
NSEREPLICAS=${NSEREPLICAS:-1}

for i in "$@"; do
    case $i in
        --nse-hub=?*)
	    NSE_HUB=${i#*=}
	    ;;
        --nse-tag=?*)
            NSE_TAG=${i#*=}
	    ;;
        -h|--help)
            usage
            exit
            ;;
        --namespace=?*)
            NAMESPACE=${i#*=}
            ;;
        --serviceName=?*)
            SERVICENAME=${i#*=}
            ;;
        --ipamPool=?*)
            IPAMPOOL=${i#*=}
            ;;
        --ipamOctet=?*)
            echo "ipamOctet is deprecated"
            IPAMOCTET=${i#*=}
            ;;
        --wcmNsrAddr=?*)
            WCM_NSRADDR=${i#*=}
            ;;
        --wcmNsrPort=?*)
            WCM_NSRPORT=${i#*=}
            ;;
        --nameserver=?*)
            NAMESERVER=${i#*=}
            ;;
        --dnszone=?*)
            DNSZONE=${i#*=}
            ;;
        --cleanup|--delete)
            INSTALL_OP=delete
            ;;
        --hello)
            HELLO=true
            ;;
        --nowait)
            NOWAIT=true
            ;;
        *)
            print_usage
            exit 1
            ;;
    esac
done

sdir=$(dirname ${0})
#echo "$sdir"

if [[ -n ${WCM_NSRADDR} ]]; then
    REMOTE_IP=${WCM_NSRADDR}
fi
#if [[ -n ${WCM_NSRPORT} ]]; then
#    REMOTE_IP=${REMOTE_IP}:${WCM_NSRPORT}
#fi

VL3HELMDIR=${VL3HELMDIR:-${sdir}/../../deployments/helm}

MFSTDIR=${MFSTDIR:-${sdir}/../k8s}
VL3_NSEMFST=${MFSTDIR}/vl3-nse-ucnf-single.yaml
if [[ -n ${REMOTE_IP} ]]; then
   VL3_NSEMFST=${MFSTDIR}/vl3-nse-ucnf_deploy.yaml
fi

KUBEINSTALL="kubectl $INSTALL_OP ${KCONF:+--kubeconfig $KCONF}"

CFGMAP="configmap nsm-vl3-${SERVICENAME}"
if [[ "${INSTALL_OP}" == "delete" ]]; then
    echo "Delete configmap"
    kubectl delete --namespace ${NAMESPACE} ${KCONF:+--kubeconfig $KCONF} ${CFGMAP}
else
    wcm_namespace_status=$(kubectl get namespace $NAMESPACE ${KCONF:+--kubeconfig $KCONF} -o=jsonpath='{.status.phase}')
    if [[ "${wcm_namespace_status}" == "Active" ]]; then
      echo "Namespace " ${NAMESPACE} " already exists"
    else
      kubectl create namespace ${NAMESPACE} ${KCONF:+--kubeconfig $KCONF}
    fi

    if [[ -n ${REMOTE_IP} ]]; then
        kubectl create --namespace ${NAMESPACE} ${KCONF:+--kubeconfig $KCONF} ${CFGMAP} --from-literal=remote.ip_list=${REMOTE_IP}
    fi
fi

echo "---------------Install NSE-------------"
# ${KUBEINSTALL} -f ${VL3_NSEMFST}
helm template ${VL3HELMDIR}/vl3 --set org=${NSE_HUB} --set tag=${NSE_TAG} --set pullPolicy=${PULLPOLICY} --set nsm.serviceName=${SERVICENAME} ${IPAMPOOL:+ --set nseControl.ipam.defaultPrefixPool=${IPAMPOOL}} --set nseControl.nsr.addr=${WCM_NSRADDR} ${WCM_NSRPORT:+ --set nseControl.nsr.port=${WCM_NSRPORT}} --set replicaCount=${NSEREPLICAS} ${IPAMOCTET:+--set ipamUniqueOctet=${IPAMOCTET}} --namespace=${NAMESPACE} ${NAMESERVER:+ --set nseControl.nameserver=${NAMESERVER}} ${DNSZONE:+ --set nseControl.dnszone=${DNSZONE}} | kubectl ${INSTALL_OP} ${KCONF:+--kubeconfig $KCONF} -f -

if [[ "$INSTALL_OP" != "delete" ]]; then
  sleep 20
  kubectl wait ${KCONF:+--kubeconfig $KCONF} --namespace=${NAMESPACE} --timeout=150s --for condition=Ready -l networkservicemesh.io/app=vl3-nse-${SERVICENAME} pod
fi

if [[ "${HELLO}" == "true" ]]; then
    echo "---------------Install hello-------------"
    #${KUBEINSTALL} -f ${MFSTDIR}/vl3-hello.yaml
    ${KUBEINSTALL} -f ${MFSTDIR}/vl3-hello-kali.yaml

    if [[ "$INSTALL_OP" != "delete" ]]; then
        sleep 10
        kubectl wait ${KCONF:+--kubeconfig $KCONF} --timeout=150s --for condition=Ready -l app=helloworld pod
    fi
fi
