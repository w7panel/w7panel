kubectl get resourcequotas --all-namespaces -o json | jq -r '
  .items[] | 
  .metadata.namespace as $ns |
  .status.hard | 
  objects | 
  to_entries[] | 
  select(.key | endswith("limits.cpu") or endswith("limits.memory")) |
  "\($ns) \(.key) \(.value)"
' | awk '
{
    namespace = $1
    resource = $2
    value = $3
    
    # 处理CPU limits
    if(resource == "limits.cpu") {
        if(value ~ /m$/) {
            # 转换 millicores 到 cores
            cpu = substr(value, 1, length(value)-1) / 1000
        } else {
            cpu = value
        }
        total_cpu += cpu
        ns_cpu[namespace] += cpu
    }
    
    # 处理Memory limits
    if(resource == "limits.memory") {
        # 转换内存单位到GiB
        if(value ~ /Ki$/) {
            mem = substr(value, 1, length(value)-2) / 1024 / 1024
        } else if(value ~ /Mi$/) {
            mem = substr(value, 1, length(value)-2) / 1024
        } else if(value ~ /Gi$/) {
            mem = substr(value, 1, length(value)-2)
        } else if(value ~ /Ti$/) {
            mem = substr(value, 1, length(value)-2) * 1024
        } else {
            # 假设是bytes
            mem = value / 1024 / 1024 / 1024
        }
        total_memory += mem
        ns_memory[namespace] += mem
    }
}
END {
    print "=== CPU Limits Summary ==="
    for(ns in ns_cpu) {
        printf "Namespace: %-20s CPU Limits: %.2f cores\n", ns, ns_cpu[ns]
    }
    printf "%-35s %.2f cores\n\n", "Total CPU Limits:", total_cpu
    
    print "=== Memory Limits Summary ==="
    for(ns in ns_memory) {
        printf "Namespace: %-20s Memory Limits: %.2f GiB\n", ns, ns_memory[ns]
    }
    printf "%-35s %.2f GiB\n", "Total Memory Limits:", total_memory
}'