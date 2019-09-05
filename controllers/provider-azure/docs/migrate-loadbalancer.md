# Migrate Azure Shoot LoadBalancer from basic to standard SKU

This guide descibes how to migrate the LoadBalancer of a Azure Shoot cluster from the basic SKU to the standard SKU.<br/>
**Be aware:** You need to delete and recreate all services of type LoadBalancer, which means that the public ip addresses of your service endpoints will change.<br/>
Please do this only if you really need to migrate this Shoot to use standard LoadBalancers. All new Shoot cluster will automatically use standard LoadBalances.

1. Disable temporarily Gardeners reconciliation.
```sh
# In the Garden cluster.
kubectl annotate shoot <shoot-name> shoot.garden.sapcloud.io/ignore="true"

# In the Seed cluster.
kubectl -n <shoot-namespace> scale deployment gardener-resource-manager --replicas=0
```

2. Backup all Kubernetes services of type LoadBalancer.
```sh
# In the Shoot cluster.
# Determine all LoadBalancer services.
kubectl get service --all-namespaces | grep LoadBalancer

# Backup each LoadBalancer service.
echo "---" >> service-backup.yaml && kubectl -n <namespace> get service <service-name> -o yaml >> service-backup.yaml
```

3. Delete all LoadBalancer services.
```sh
# In the Shoot cluster.
kubectl -n <namespace> delete service <service-name>
```

4. Wait until until LoadBalancer is deleted.
Wait until all services of type LoadBalancer are deleted and the LoadBalancer on Azure is also deleted.
Check via the Azure Portal if the LoadBalancer within the Shoot Resource Group has been deleted.
This should happen automatically after all Kubernetes LoadBalancer service are gone within a few minutes.

5. Modify the `cloud-povider-config` configmap in the Seed namespace of the Shoot.<br/>
The key `cloudprovider.conf` contains the Kubernetes cloud-provider configuration.
The value is a multiline string. Please append the this `loadBalancerSku: \"standard\"\n` to the value/string.

```sh
# In the Seed cluster.
kubectl -n <shoot-namespace> edit cm cloud-provider-config
```

6. Enable Gardeners reconcilation and trigger a reconciliation.
```
# In the Garden cluster
# Enable reconcilation
kubectl annotate shoot <shoot-name> shoot.garden.sapcloud.io/ignore-

# Trigger reconcilation
kubectl annotate shoot <shoot-name> shoot.garden.sapcloud.io/operation="reconcile"
```
Wait until the cluster has been reconciled.

6. Recreate the services from the backup file.<br/>
Probably you need to remove some fields from the service defintions e.g. `.spec.clusterIP`, `.metadata.uid` or `.status` etc.
```sh
kubectl apply -f service-backup.yaml
```

7. If successful remove backup file.
```sh
# Delete the backup file.
rm -f service-backup.yaml
```

