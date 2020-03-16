#!/bin/bash

kubectl delete -f deployment/deployment-latest.yaml && \
  kubectl delete secret lazykube-webhook-certs && \
  kubectl delete mutatingwebhookconfiguration lazykube-webhook-cfg

kubectl delete pods -l lazykubetest=e2e-test

kubectl delete -f e2e/dashboard.yaml