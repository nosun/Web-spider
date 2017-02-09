package middleware

import (
	"math"
	"sync"
)

type myIDGenertor struct {
	sync  sync.Mutex
	ID    uint32
	ended bool //前一个数是否最大
}

func NewIDGenertor() IDGenertor {
	return &myIDGenertor{}
}

//todo 越界重置为0.极限43亿次 足够？
func (IDg *myIDGenertor) GetUint32() uint32 {
	IDg.sync.Lock()
	defer IDg.sync.Unlock()
	if IDg.ended {
		defer func() {
			IDg.ended = false
		}()
		IDg.ID = 0
		return IDg.ID
	}
	if IDg.ID < math.MaxUint32 {
		IDg.ID++
	} else {
		IDg.ended = true
	}
	return IDg.ID
}
