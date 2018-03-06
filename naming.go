package tafgo

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var DefaultNamingService *QueryFProxy
var namingMap sync.Map
var routineTaskLaunched int32

func NewDefaultNaming(obj string, timeout time.Duration) {
	DefaultNamingService = NewQueryFProxy(obj, timeout)
	if atomic.CompareAndSwapInt32(&routineTaskLaunched, 0, 1) {
		go func() {
			routineObjs()
		}()
	}
}

func getEndpoints(addr string) []EndpointF {
	v, exist := namingMap.Load(addr)
	if exist {
		return v.([]EndpointF)
	}
	return nil
}
func saveEndpoints(addr string, v []EndpointF) {
	namingMap.Store(addr, v)
}

func loadEndpointByName(addr string) {
	endpoints, _, err := DefaultNamingService.FindObjectById(addr, nil)
	if nil != err {
		log.Printf("ERROR:Failed to FindObjectById with name:%v and error:%v", addr, err)
	} else {
		namingMap.Store(addr, endpoints)
	}
}

func routineObjs() {
	for {
		select {
		case <-time.After(30 * time.Second):
			namingMap.Range(func(key, value interface{}) bool {
				loadEndpointByName(key.(string))
				return true
			})
		}
	}
}

func addWatchObj(obj string) {
	var empty []EndpointF
	namingMap.Store(obj, empty)
	loadEndpointByName(obj)
}
