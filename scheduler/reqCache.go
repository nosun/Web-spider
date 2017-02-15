package scheduler

import (
	"github.com/yanchenxu/Web-spider/base"
	"sync"
)

type myRequest struct {
	cache  []base.Request
	mutexx sync.Mutex
	status byte //0 正常运行  1  关闭
}

func newRequestCache() requestCache {
	return nil
}
