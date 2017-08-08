package main

import "sync"

type HostConnNumMap struct {
	m    map[string]int
	lock sync.Mutex
}

func (HostConnNumMap *HostConnNumMap) AddCount(host string) {
	HostConnNumMap.lock.Lock()
	defer HostConnNumMap.lock.Unlock()
	HostConnNumMap.m[host] += 1
}

func (HostConnNumMap *HostConnNumMap) DoneCount(host string) {
	HostConnNumMap.lock.Lock()
	defer HostConnNumMap.lock.Unlock()
	HostConnNumMap.m[host] -= 1
}

func (HostConnNumMap *HostConnNumMap) GetCount(host string) int {
	HostConnNumMap.lock.Lock()
	defer HostConnNumMap.lock.Unlock()
	return HostConnNumMap.m[host]
}

func (HostConnNumMap *HostConnNumMap) DelHost(host string) {
	HostConnNumMap.lock.Lock()
	defer HostConnNumMap.lock.Unlock()
	delete(HostConnNumMap.m, host)
}
