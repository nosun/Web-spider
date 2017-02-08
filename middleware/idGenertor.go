package middleware

import "sync"

type myIDGenertor struct {
	sync sync.Mutex
	ID   uint32
}

func NewIDGenertor() IDGenertor {
	return &myIDGenertor{}
}

//todo 越界？
func (IDg *myIDGenertor) GetUint32() uint32 {
	IDg.sync.Lock()
	defer IDg.sync.Unlock()
	IDg.ID++
	return IDg.ID
}
