#!/bin/bash

KFCTL_URL="https://github.com/kubeflow/kfctl/releases/download/v1.0/kfctl_v1.0-0-g94c35cf_linux.tar.gz"
DIR=$(mktemp -d)
echo "temp path: ${DIR}"
curl -x 127.0.0.1:8118 -LJ -o $DIR/kfctl.tar.gz $KFCTL_URL
cd $DIR
tar -xvf kfctl.tar.gz

KF_NAME="kubeflow-test"
BASE_DIR="${DIR}/mydep"
KF_DIR=${BASE_DIR}/${KF_NAME}
CONFIG_URI="https://raw.githubusercontent.com/kubeflow/manifests/v1.0-branch/kfdef/kfctl_k8s_istio.v1.0.0.yaml"
mkdir -p ${KF_DIR}
cd ${KF_DIR}
${DIR}/kfctl apply -V -f ${CONFIG_URI}

# cd ${KF_DIR}
# # If you want to delete all the resources, run:
# kfctl delete -f ${CONFIG_FILE}