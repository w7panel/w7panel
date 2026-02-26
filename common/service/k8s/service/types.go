package service

type void struct{}
type allowSvcProxy map[string]map[string]void

var allowSvcProxyMap = make(allowSvcProxy)

func (a allowSvcProxy) appendAllowProxy(ns, name string) {
	if a == nil {
		return
	}
	if _, ok := a[ns]; !ok {
		a[ns] = make(map[string]void)
	}
	a[ns][name] = void{}
}
func (a allowSvcProxy) removeAllowProxy(ns, name string) {
	if a == nil {
		return
	}
	delete(a[ns], name)
}
func (a allowSvcProxy) IsAllowProxy(ns, name string) bool {
	if a == nil {
		return false
	}
	_, ok := a[ns][name]
	return ok
}

func IsAllowProxy(ns, name string) bool {
	return allowSvcProxyMap.IsAllowProxy(ns, name)
}
