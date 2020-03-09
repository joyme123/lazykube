#!/bin/bash

./deployment/webhook-create-signed-cert.sh \
    --service lazykube-webhook-svc \
    --secret lazykube-webhook-certs \
    --namespace default

cat deployment/mutatingwebhook.yaml | \
    deployment/webhook-patch-ca-bundle.sh > \
    deployment/mutatingwebhook-ca-bundle.yaml

kubectl create -f deployment/deployment.yaml
kubectl create -f deployment/mutatingwebhook-ca-bundle.yaml

# 测试创建pod
kubectl create -f e2e/gcr.io-pod.yaml
kubectl create -f e2e/quayio-pod.yaml
kubectl create -f e2e/k8s.gcr.io-pod.yaml

# 测试kubeflow
e2e/kubeflow/deploy.sh
