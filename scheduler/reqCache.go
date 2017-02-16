package scheduler

import (
	"github.com/yanchenxu/Web-spider/base"
	"sync"
	"fmt"
)

type myRequestCache struct {
	cache  []*base.Request
	mutexx sync.Mutex
	status byte //0 正常运行  1  关闭
}

func newRequestCache() requestCache {
	return &myRequestCache{
		cache:make([]*base.Request, 0),
	}
}

func (reqCache *myRequestCache)put(req *base.Request) bool {
	if req == nil {
		return false
	}

	if reqCache.status == 1 {
		return false
	}

	reqCache.mutexx.Lock()
	defer reqCache.mutexx.Unlock()

	reqCache.cache = append(reqCache.cache, req)

	return true
}

func (reqCache *myRequestCache)get() *base.Request {
	if reqCache.length() == 0 {
		return nil
	}
	if reqCache.status == 1 {
		return nil
	}

	reqCache.mutexx.Lock()
	defer reqCache.mutexx.Unlock()

	req := reqCache.cache[0]
	reqCache.cache = reqCache.cache[1:]
	return req
}

func (reqCache *myRequestCache)capacity() int {
	return cap(reqCache.cache)
}

func (reqCache *myRequestCache)length() int {
	return len(reqCache.cache)
}

func (reqCache *myRequestCache)close() {
	if reqCache.status == 1 {
		return
	}
	reqCache.status = 1
}

//摘要信息模板
var summaryTemplate = "status: %s," + "length: %d," + "capacity: %d"
//状态字典
var statusMap = map[byte]string{
	0:"running",
	1:"closed",
}

func (reqCache *myRequestCache)summary() string {
	return fmt.Sprintf(summaryTemplate, statusMap[reqCache.status], reqCache.length(), reqCache.capacity())
}

