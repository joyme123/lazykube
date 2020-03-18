#!/bin/bash

# 测试安装
./deployment/webhook-create-signed-cert.sh \
    --service lazykube-webhook-svc \
    --secret lazykube-webhook-certs \
    --namespace kube-system >> /dev/null

cat deployment/mutatingwebhook.yaml | \
    deployment/webhook-patch-ca-bundle.sh > \
    deployment/mutatingwebhook-ca-bundle.yaml

kubectl create -f deployment/deployment-latest.yaml >> /dev/null && \
    kubectl create -f deployment/mutatingwebhook-ca-bundle.yaml >> /dev/null

# 等待 lazykube 启动

while true
do

if [[ $(kubectl get pods  -l app=lazykube | grep Running) != "" ]]
then
    echo "lazykube 已经启动"
    break
else
    echo "等待 lazykube 启动"
    sleep 2
fi
done

# 测试创建 gcr.io pod
kubectl create -f e2e/gcr.io-pod.yaml >> /dev/null
image=$(kubectl get pods myapp-gcr -o=jsonpath='{.spec.initContainers[0].image}')
expect="gcr.azk8s.cn/kubernetes-helm/tiller:v2.13.1"
if [[ $image != "$expect" ]]
then
    echo "gcr.io pod initContainers test failed, result is ${image}, expect is ${expect}"
    exit 1
fi

image=$(kubectl get pods myapp-gcr -o=jsonpath='{.spec.containers[0].image}')
expect="gcr.azk8s.cn/kubernetes-helm/tiller:v2.13.1"
if [[ $image != "$expect" ]]
then
    echo "gcr.io pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/gcr.io-pod.yaml >> /dev/null

# 测试创建 quay.io pod
kubectl create -f e2e/quayio-pod.yaml >> /dev/null
image=$(kubectl get pods myapp-quay -o=jsonpath='{.spec.initContainers[0].image}')
expect="quay.azk8s.cn/dexidp/dex:v2.10.0"
if [[ $image != "$expect" ]]
then
    echo "quay.io pod initContainers test failed, result is ${image}, expect is ${expect}"
    exit 1
fi

image=$(kubectl get pods myapp-quay -o=jsonpath='{.spec.containers[0].image}')
expect="quay.azk8s.cn/dexidp/dex:v2.10.0"
if [[ $image != "$expect" ]]
then
    echo "quay.io pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/quayio-pod.yaml >> /dev/null

# 测试创建 k8s.gcr.io
kubectl create -f e2e/k8s.gcr.io-pod.yaml >> /dev/null
image=$(kubectl get pods myapp-k8s-gcr -o=jsonpath='{.spec.containers[0].image}')
expect="gcr.azk8s.cn/google-containers/addon-resizer:1.8.4"
if [[ $image != "$expect" ]]
then
    echo "k8s.gcr.io pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
image=$(kubectl get pods myapp-k8s-gcr -o=jsonpath='{.spec.containers[1].image}')
expect="gcr.azk8s.cn/google-containers/addon-resizer:1.8.4"
if [[ $image != "$expect" ]]
then
    echo "k8s.gcr.io pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/k8s.gcr.io-pod.yaml >> /dev/null

# 测试创建 docker.io
kubectl create -f e2e/docker.io-pod.yaml >> /dev/null
image=$(kubectl get pods myapp-docker -o=jsonpath='{.spec.containers[0].image}')
expect="dockerhub.azk8s.cn/library/mysql:5.6"
if [[ $image != "$expect" ]]
then
    echo "docker.io pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/docker.io-pod.yaml >> /dev/null

# 测试 mysql:5.6 这种格式的镜像

kubectl create -f e2e/default-pod.yaml >> /dev/null
image=$(kubectl get pods myapp-docker-default -o=jsonpath='{.spec.containers[0].image}')
expect="dockerhub.azk8s.cn/library/mysql:5.6"
if [[ $image != "$expect" ]]
then
    echo "default pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/default-pod.yaml >> /dev/null

# 测试 library/mysql:5.6 这种格式的镜像

kubectl create -f e2e/default-pod-2.yaml >> /dev/null
image=$(kubectl get pods myapp-docker-default-2 -o=jsonpath='{.spec.containers[0].image}')
expect="dockerhub.azk8s.cn/library/mysql:5.6"
if [[ $image != "$expect" ]]
then
    echo "default 2 pod container test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/default-pod-2.yaml >> /dev/null

# 测试dashboard
kubectl create -f e2e/dashboard.yaml >> /dev/null
image=$(kubectl -n kube-system get pods -l k8s-app=kubernetes-dashboard -o=jsonpath='{.items[0].spec.containers[0].image}')
expect="gcr.azk8s.cn/google-containers/kubernetes-dashboard-amd64:v1.10.1"
if [[ $image != "$expect" ]]
then
    echo "dashboard test failed, result is ${image}, expect is ${expect}"
    exit 1
fi
kubectl delete -f e2e/dashboard.yaml >> /dev/null

# 测试卸载 lazykube
kubectl delete -f deployment/deployment-latest.yaml >> /dev/null && \
  kubectl -n kube-system delete secret lazykube-webhook-certs >> /dev/null && \
  kubectl delete mutatingwebhookconfiguration lazykube-webhook-cfg >> /dev/null