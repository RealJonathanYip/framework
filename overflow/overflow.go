package overflow

import (
	"sync"
	"sync/atomic"
	"time"
)

type OverFlowConfig struct {
	CurrentQps uint32
	TimeStamp  int64
}

var (
	_key2QpsLimit map[string]*OverFlowConfig
	_rwLock       *sync.RWMutex
)

func init() {
	_key2QpsLimit = make(map[string]*OverFlowConfig)
	_rwLock = new(sync.RWMutex)
}

func IsOverFlow(szKey string, nQps uint32) bool {
	_rwLock.RLock()
	ptrConfig, bExist := _key2QpsLimit[szKey]
	_rwLock.RUnlock()

	objTime := time.Now()
	if !bExist {
		ptrConfig = new(OverFlowConfig)
		ptrConfig.CurrentQps = 0
		ptrConfig.TimeStamp = objTime.Unix()

		_rwLock.Lock()
		_key2QpsLimit[szKey] = ptrConfig
		_rwLock.Unlock()
	}

	if atomic.LoadInt64(&ptrConfig.TimeStamp) != objTime.Unix() {
		atomic.StoreInt64(&ptrConfig.TimeStamp, objTime.Unix())
		atomic.StoreUint32(&ptrConfig.CurrentQps, 1)
		return false
	} else if atomic.LoadUint32(&ptrConfig.CurrentQps) >= nQps {
		return true
	} else {
		atomic.AddUint32(&ptrConfig.CurrentQps, 1)
		return false
	}
}
