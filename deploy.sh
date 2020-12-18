#!/bin/bash

set -e
trap 'catch $? $LINENO' EXIT

catch() {
    if [ "$1" != "0" ]; then
        echo "Error $1 occurred on line $2"
    fi
}

# You might need 'brew install hudochenkov/sshpass/sshpass'
if [ ! -f master-config ]; then
    sshpass -p 'initial0' scp root@rabbit-dev-vm-k8s-master:~/.kube/config ./master-config
fi

export KUBECONFIG=`pwd`/master-config

kubectl config view
kubectl config use-context kubernetes-admin@kubernetes

helm del --purge nnf-ec
