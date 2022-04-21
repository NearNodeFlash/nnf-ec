#!/bin/bash

# Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
# Other additional copyright holders may be indicated within.
#
# The entirety of this work is licensed under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
#
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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

helm del --purge nnf-ec || true
helm install --name nnf-ec ./kubernetes/nnf-ec/

