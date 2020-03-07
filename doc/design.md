# 设计文档

## 目标

针对在 kubernetes 集群中，国内无法下载 gcr.io 和 quay.io 镜像的问题，提出的一种无需翻墙，自动下载镜像的方案。

## 方案

使用 DaemonSet 将程序部署到 kubernetes 集群的每个节点上，watch pod 在集群上的调度，然后使用 hostpath 来访问 docker.sock，和节点上的 dockerd 通信，根据指定的镜像替换规则（具体的规则如下）来　pull 并重新 tag 镜像。

## 镜像替换规则

