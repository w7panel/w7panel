package core

type K3kUserMode string

const (
	K3kUserModeNormal  K3kUserMode = "normal"  //普通模式
	K3kUserModeFounder K3kUserMode = "founder" //创始人模式
	K3kUserModeCluster K3kUserMode = "cluster" //集群模式
)
