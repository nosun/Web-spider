package middleware

import "sync"

type IDGenertor struct {
	sync sync.Mutex
	ID   uint32
}

func NewIDGenertor() *IDGenertor {
	return IDGenertor{}
}
//todo 越界？
func (IDg *IDGenertor)GetUint32() uint32 {
	IDg.sync.Lock()
	defer IDg.sync.Unlock()
	IDg.ID++
	return IDg.ID
}
