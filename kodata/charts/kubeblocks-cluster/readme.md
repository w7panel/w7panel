pgsql

helm install pgsql-helm ./kubeblocks-cluster --set cluster.clusterDefinitionRef="postgresql" --set cluster.componentDefRef=postgresql --set cluster.clusterVersionRef=postgresql-16.4.0 --set cluster.storageClassName="disk1" 


helm install redis-helm ./kubeblocks-cluster --set cluster.clusterDefinitionRef="redis" --set cluster.componentDefRef=redis --set cluster.clusterVersionRef=redis-7.2.4 --set cluster.storageClassName="disk1" 