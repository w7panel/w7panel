#!/bin/bash


helm install mysql-80 ./kubeblocks-cluster  --set cluster.serviceVersion=8.0.33 --set cluster.mysql=true

helm install pgsql-15 ./kubeblocks-cluster --set cluster.componentDef=postgresql-15 --set cluster.postgresql=true

helm install redis-71 ./kubeblocks-cluster --set cluster.serviceVersion=7.0.6 --set cluster.redis=true

helm install mongodb-60 ./kubeblocks-cluster --set cluster.serviceVersion=6.0.16 --set cluster.mongodb=true

# helm install pgsql-cli-12-14 ./kubeblocks-cluster --set cluster.postgresql=true --set cluster.componentDef="postgresql-12" --set cluster.serviceVersion="12.14.0"