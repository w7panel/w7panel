package webdav

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"

	"gitee.com/we7coreteam/k8s-offline/common/service/procpath"
)

// UserGroupCache 全局 UserGroup 缓存管理器
type UserGroupCache struct {
	cache map[string]*UserGroup
	mu    sync.RWMutex
}

var (
	// 全局缓存实例
	globalUserGroupCache = &UserGroupCache{
		cache: make(map[string]*UserGroup),
	}
)

// GetUserGroup 获取或创建 UserGroup 实例（带缓存）
// 如果缓存中存在且进程仍然存活，返回缓存的实例
// 如果进程已销毁，删除缓存并创建新的实例
func GetUserGroup(pid string) *UserGroup {
	globalUserGroupCache.mu.Lock()
	defer globalUserGroupCache.mu.Unlock()

	// 检查缓存中是否存在
	if ug, exists := globalUserGroupCache.cache[pid]; exists {
		// 检查进程是否仍然存活（通过检查 /proc/{pid} 目录是否存在）
		if isProcessAlive(pid) {
			return ug
		}
		// 进程已销毁，删除缓存
		delete(globalUserGroupCache.cache, pid)
	}

	// 创建新的 UserGroup 实例并缓存
	ug := NewUserGroup(pid)
	globalUserGroupCache.cache[pid] = ug
	return ug
}

func isProcessAlive(pid string) bool {
	return procpath.IsProcessAlive(pid)
}

// ClearUserGroupCache 手动清除指定 PID 的缓存（用于测试或特殊场景）
func ClearUserGroupCache(pid string) {
	globalUserGroupCache.mu.Lock()
	defer globalUserGroupCache.mu.Unlock()
	delete(globalUserGroupCache.cache, pid)
}

// ClearAllUserGroupCache 清除所有缓存
func ClearAllUserGroupCache() {
	globalUserGroupCache.mu.Lock()
	defer globalUserGroupCache.mu.Unlock()
	globalUserGroupCache.cache = make(map[string]*UserGroup)
}

// NewUserGroup 创建新的 UserGroup 实例（不缓存）
// 一般情况下应使用 GetUserGroup 以利用缓存
func NewUserGroup(pid string) *UserGroup {
	ug := &UserGroup{
		pid: pid,
	}
	return ug
}

type UserGroup struct {
	pid      string
	users    map[int]string
	groups   map[int]string
	initOnce sync.Once
	mu       sync.RWMutex
}

func (ug *UserGroup) init() error {
	var initErr error
	ug.initOnce.Do(func() {
		ug.mu.Lock()
		defer ug.mu.Unlock()

		ug.users = make(map[int]string)
		ug.groups = make(map[int]string)

		// 读取passwd文件
		if err := ug.parsePasswdFile(); err != nil {
			initErr = err
			return
		}

		// 读取group文件
		if err := ug.parseGroupFile(); err != nil {
			initErr = err
			return
		}
	})
	return initErr
}

func (ug *UserGroup) parsePasswdFile() error {
	passwdPath := procpath.GetEtcPasswdPath(ug.pid)
	file, err := os.Open(passwdPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			uid, err := strconv.Atoi(parts[2])
			if err == nil {
				ug.users[uid] = parts[0]
			}
		}
	}
	return scanner.Err()
}

func (ug *UserGroup) parseGroupFile() error {
	groupPath := procpath.GetEtcGroupPath(ug.pid)
	file, err := os.Open(groupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			gid, err := strconv.Atoi(parts[2])
			if err == nil {
				ug.groups[gid] = parts[0]
			}
		}
	}
	return scanner.Err()
}

func (ug *UserGroup) GetUserName(uid int) (string, error) {
	if err := ug.init(); err != nil {
		return "", err
	}

	ug.mu.RLock()
	defer ug.mu.RUnlock()

	if name, ok := ug.users[uid]; ok {
		return name, nil
	}
	return strconv.Itoa(uid), nil
}

func (ug *UserGroup) GetGroupName(gid int) (string, error) {
	if err := ug.init(); err != nil {
		return "", err
	}

	ug.mu.RLock()
	defer ug.mu.RUnlock()

	if name, ok := ug.groups[gid]; ok {
		return name, nil
	}
	return strconv.Itoa(gid), nil
}

// GetAllUsers 获取所有用户列表（用于前端权限设置时的下拉选择）
func (ug *UserGroup) GetAllUsers() ([]map[string]interface{}, error) {
	if err := ug.init(); err != nil {
		return nil, err
	}

	ug.mu.RLock()
	defer ug.mu.RUnlock()

	var users []map[string]interface{}
	for uid, name := range ug.users {
		users = append(users, map[string]interface{}{
			"id":   uid,
			"name": name,
		})
	}
	return users, nil
}

// GetAllGroups 获取所有组列表
func (ug *UserGroup) GetAllGroups() ([]map[string]interface{}, error) {
	if err := ug.init(); err != nil {
		return nil, err
	}

	ug.mu.RLock()
	defer ug.mu.RUnlock()

	var groups []map[string]interface{}
	for gid, name := range ug.groups {
		groups = append(groups, map[string]interface{}{
			"id":   gid,
			"name": name,
		})
	}
	return groups, nil
}
