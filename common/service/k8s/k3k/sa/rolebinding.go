package sa

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoleBinding struct {
	client.Client
}

func NewRoleBinding(c client.Client) *RoleBinding {
	return &RoleBinding{
		Client: c,
	}
}

func (r *RoleBinding) createRoleAndBinding(ctx context.Context, name string, namespace string, saNamespace string, saName string) error {
	// 创建命名空间级别的 Role
	// 导入所需的包

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log", "limitranges", "resourcequotas", "resourcequotas/status", "limitranges/status"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"metrics.k8s.io"},
				Resources: []string{"pods"}, // 只包含命名空间级别的资源
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// Create or update Role
	if err := r.Create(ctx, role); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, role); err != nil {
				return fmt.Errorf("failed to update Role in namespace %s: %v", namespace, err)
			}
		} else {
			return fmt.Errorf("failed to create Role in namespace %s: %v", namespace, err)
		}
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: saNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Create or update RoleBinding
	if err := r.Create(ctx, roleBinding); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, roleBinding); err != nil {
				return fmt.Errorf("failed to update RoleBinding in namespace %s: %v", namespace, err)
			}
		} else {
			return fmt.Errorf("failed to create RoleBinding in namespace %s: %v", namespace, err)
		}
	}

	return nil
}

// 创建集群级别的 ClusterRole 和 ClusterRoleBinding
func (r *RoleBinding) createClusterRoleAndBinding(ctx context.Context, name string, saNamespace string, saName string) error {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "metrics-" + name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"metrics.k8s.io"},
				Resources: []string{"nodes"}, // 集群级别的资源
				Verbs:     []string{"get", "list", "watch"},
			},

			{
				APIGroups: []string{""},
				Resources: []string{"services/proxy"}, // 集群级别的资源
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// Create or update ClusterRole
	if err := r.Create(ctx, clusterRole); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, clusterRole); err != nil {
				return fmt.Errorf("failed to update ClusterRole: %v", err)
			}
		} else {
			return fmt.Errorf("failed to create ClusterRole: %v", err)
		}
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "metrics-" + name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: saNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "metrics-" + name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Create or update ClusterRoleBinding
	if err := r.Create(ctx, clusterRoleBinding); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, clusterRoleBinding); err != nil {
				return fmt.Errorf("failed to update ClusterRoleBinding: %v", err)
			}
		} else {
			return fmt.Errorf("failed to create ClusterRoleBinding: %v", err)
		}
	}

	return nil
}

func (r *RoleBinding) CreateRole(ctx context.Context, sa *v1.ServiceAccount, k3kNamespace string) error {

	// Create role and binding in sa's namespace
	if err := r.createRoleAndBinding(ctx, sa.Name, sa.Namespace, sa.Namespace, sa.Name); err != nil {
		return err
	}

	// Create role and binding in k3k namespace
	if err := r.createRoleAndBinding(ctx, sa.Name, k3kNamespace, sa.Namespace, sa.Name); err != nil {
		return err
	}

	// Create cluster role and binding for metrics.k8s.io/nodes resource
	if err := r.createClusterRoleAndBinding(ctx, sa.Name, sa.Namespace, sa.Name); err != nil {
		return err
	}

	return nil
}

func (r *RoleBinding) CreateNormalUserRoleBinding(ctx context.Context, sa *v1.ServiceAccount, role string) error {

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "w7panel-rb-" + role,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     role,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Create or update ClusterRoleBinding
	if err := r.Create(ctx, clusterRoleBinding); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, clusterRoleBinding); err != nil {
				return fmt.Errorf("failed to update ClusterRoleBinding: %v", err)
			}
		} else {
			return fmt.Errorf("failed to create ClusterRoleBinding: %v", err)
		}
	}

	return nil
}
