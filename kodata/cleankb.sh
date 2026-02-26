#!/bin/sh




kubectl get storageproviders.dataprotection.kubeblocks.io -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get components -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get configconstraints -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get configurations -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -

kubectl get clusterversion -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get componentclassdefinitions -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get componentdefinitions -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get opsdefinitions -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get componentversions -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get clusters -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -

kubectl get backupschedules -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get restores -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get backuppolicies -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get actionsets -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
kubectl get ops -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
# 2ge storageproviders
kubectl get storageproviders.storage.kubeblocks.io -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -

kubectl get opsdefinitions -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -

kubectl get configmap -o json | jq '.items[] | select(.metadata.finalizers != null) | .metadata.finalizers |= map(select(. == "config.kubeblocks.io/finalizer"))' | kubectl apply -f -
kubectl get configmap -o json | jq '.items[] | select(.metadata.finalizers != null) | .metadata.finalizers |= map(select(. == "component.kubeblocks.io/finalizer"))' | kubectl apply -f -
# kubectl get configmap -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -

#//component.kubeblocks.io/finalizer

kubectl get secrets -o json | jq '.items[] | select(.metadata.finalizers != null) | .metadata.finalizers |= map(select(. == "config.kubeblocks.io/finalizer"))' | kubectl apply -f -
kubectl get secrets -o json | jq '.items[] | select(.metadata.finalizers != null) | .metadata.finalizers |= map(select(. == "component.kubeblocks.io/finalizer"))' | kubectl apply -f -
# kubectl get secrets -o json | jq '.items[] | .metadata.finalizers = []' | kubectl apply -f -
