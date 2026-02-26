package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserSpec `json:"spec"`
}

type UserSpec struct {
	// Username string `json:"username"` // 用户名就是 metadata.name
	IsSuper    bool     `json:"isSuper,omitempty"`    // 是否为超级管理员
	Password   string   `json:"password,omitempty"`   // 密码可以为空，如果不为空则在创建时就设置密码
	Namespaces []string `json:"namespaces,omitempty"` // 用户可以访问的命名空间列表
	Groups     []string `json:"groups,omitempty"`     // 用户所属的用户组列表
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []User `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type UserGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec UserGroupSpec `json:"spec"`
}

type UserGroupSpec struct {
	// Username string `json:"username"` // 用户名就是 metadata.name
	Password   string   `json:"password,omitempty"`   // 密码可以为空，如果不为空则在创建时就设置密码
	Namespaces []string `json:"namespaces,omitempty"` // 用户组可以访问的命名空间列表
	Users      []string `json:"users,omitempty"`      // 用户组包含的用户列表
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type UserGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []UserGroup `json:"items"`
}
