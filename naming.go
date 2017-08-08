package tafgo

import (
	"time"
)

var DefaultNamingService *QueryFProxy

func NewDefaultNaming(obj string, timeout time.Duration) {
	DefaultNamingService = NewQueryFProxy(obj, timeout)
}
