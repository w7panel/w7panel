#!/bin/bash
#https://www.baeldung-cn.com/linux/bcrypt-hash
# 显示菜单
show_menu() {
    echo "欢迎使用宝塔命令行工具"
    echo "请选择一个选项 (1-12):"
    echo "1. 查看系统信息"
    echo "2. 重启k3s"
    echo "3. 检查服务是否运行"
    echo "4. 修改用户名密码"
    echo "5. 更新应用镜像"
    echo "12. 退出"
}

checkPodRunning() {
    label="app=k8s-offline"
    pod_names=$(kubectl get pods -l "$label" -o jsonpath='{.items[*].metadata.name}')

if [ -z "$pod_names" ]; then
  echo "未找到标签为 $label 的 Pod。"
  exit 1
fi

# 检查每个 Pod 的状态
for pod_name in $pod_names; do
  pod_status=$(kubectl get pod "$pod_name" -o jsonpath='{.status.phase}')
  pod_conditions=$(kubectl get pod "$pod_name" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')

  if [ "$pod_status" == "Running" ] && [ "$pod_conditions" == "True" ]; then
    echo "Pod $pod_name 运行成功。"
  else
    echo "Pod $pod_name 运行失败。状态: $pod_status，就绪状态: $pod_conditions。"
  fi
done
}

updateImage() {
    label="app=k8s-offline"
    new_image="ccr.ccs.tencentyun.com/afan-public/k8s-offline:0.1.77"
    deployment_names=$(kubectl get deployment -l "$label" -o jsonpath='{.items[*].metadata.name}')

if [ -z "$deployment_names" ]; then
  echo "未找到标签为 $label 的 Deployment。"
  exit 1
fi

# 更新每个 Deployment 的镜像
for deployment_name in $deployment_names; do
  echo "正在更新 Deployment $deployment_name 的镜像为 $new_image..."
  kubectl set image deployment "$deployment_name" *="$new_image"

  if [ $? -eq 0 ]; then
    echo "更新成功。"
  else
    echo "应用 $deployment_name 更新失败。"
  fi
done
}

# 根据用户选择执行相应的命令
execute_command() {
    case $1 in
        1)
            echo "正在查看系统信息..."
            uname -a
            ;;
        2)
            echo "正在重启k3s..."
            sudo systemctl restart k3s  # 假设使用Nginx
            ;;
        3)
            echo "检测服务是否运行..."
            checkPodRunning
            ;;
        4)
            echo "正在修改用户名密码..."
            read -p "请输入新用户名: " new_username
            echo "正在修改密码..."
            read -sp "请输入新密码: " new_password
            echo ""
            read -sp "请再次输入新密码: " confirm_password
            echo ""
            if [ "$new_password" == "$confirm_password" ]; then
                echo "新密码: $new_password"
                kubectl get pods -l app=k8s-offline -o name | xargs -I {} kubectl exec {} -- k8s-offline auth:register --username=$new_username --password=$confirm_password
            else
                echo "密码不匹配，请重试。"
            fi
            ;;
        5)
            echo "正在更新..."
            updateImage
            ;;
        12)
            echo "退出命令行工具"
            exit 0
            ;;
        *)
            echo "无效的选择，请输入1-12之间的数字。"
            ;;
    esac
}

# 主循环
# while true; do
    show_menu
    read -p "请输入选项: " choice
    execute_command $choice
    echo ""  # 输出空行以分隔命令输出
# done

