package middleware

import (
	"fmt"
	"sync"
)

type myStopSign struct {
	stgned       bool              //信号是否已发出
	dealCountMap map[string]uint32 //处理计数的字典
	rwmutex      sync.RWMutex      //读写锁
}

func NewStopSign() StopSign {
	return &myStopSign{dealCountMap: make(map[string]uint32)}
}

//发出停止型号，如果已经发出，返回false
func (ss *myStopSign) Sign() bool {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if ss.stgned {
		return false
	}
	ss.stgned = true
	return true
}

//判断停止型号是否发出
func (ss *myStopSign) Signed() bool {
	return ss.stgned
}

//重置
func (ss *myStopSign) Reset() {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	ss.stgned = false
	ss.dealCountMap = make(map[string]uint32)
}

//处理
func (ss *myStopSign) Deal(code string) {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if !ss.stgned {
		return
	}
	if _, ok := ss.dealCountMap[code]; !ok {
		ss.dealCountMap[code] = 1
	} else {
		ss.dealCountMap[code] += 1
	}
}

//获得处理计数
func (ss *myStopSign) DealCount(code string) uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	return ss.dealCountMap[code]
}

//总计数
func (ss *myStopSign) DealTotal() uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	var total uint32
	for _, v := range ss.dealCountMap {
		total += v
	}
	return total
}

//摘要
func (ss *myStopSign) Summary() string {
	return fmt.Sprintf("status: %d\n dealCountMap: %d\n", ss.stgned, ss.dealCountMap)
}
