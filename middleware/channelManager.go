package middleware

import (
	"fmt"
	"github.com/yanchenxu/Web-spider/base"
	"sync"
)

//被用来管理状态
type ChannelManagerStatus uint8

const (
	CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManagerStatus = 0 //未初始化
	CHANNEL_MANAGER_STATUS_INITIALIZED   ChannelManagerStatus = 1 //已初始化
	CHANNEL_MANAGER_STATUS_CLOSED        ChannelManagerStatus = 2 //已关闭
)

var statusNameMap = map[ChannelManagerStatus]string{
	CHANNEL_MANAGER_STATUS_UNINITIALIZED: "uninitizlized",
	CHANNEL_MANAGER_STATUS_INITIALIZED:   "initialized",
	CHANNEL_MANAGER_STATUS_CLOSED:        "closed",
}

//信息模板
var chanmanSummaryTemplate = "status" +
	"requestChannel" +
	"responsechanel" +
	"itemChannel" +
	"errorChannel"

const defaultChanlen uint = 10

type myChannelManager struct {
	channelAgrs base.ChannelArgs
	reqCh       chan base.Request
	respCh      chan base.Response
	itemCh      chan base.Item
	errCh       chan error
	status      ChannelManagerStatus
	rwmutex     sync.RWMutex //读写锁
}

func NewChannelManager(channelArgs base.ChannelArgs) ChannelManager {
	chanman := &myChannelManager{}
	chanman.Init(channelArgs, true)
	return chanman
}

func (chanman *myChannelManager) Init(channelArgs base.ChannelArgs, rest bool) bool {
	if err := channelArgs.Check(); err != nil {
		panic(err)
	}
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if chanman.status == CHANNEL_MANAGER_STATUS_INITIALIZED && !rest {
		return false
	}
	chanman.channelAgrs = channelArgs
	chanman.reqCh = make(chan base.Request, chanman.channelAgrs.ReqChanLen())
	chanman.respCh = make(chan base.Response, chanman.channelAgrs.RespChanLen())
	chanman.itemCh = make(chan base.Item, chanman.channelAgrs.ItemChanLen())
	chanman.errCh = make(chan error, chanman.channelAgrs.ErrorChanLen())
	chanman.status = CHANNEL_MANAGER_STATUS_INITIALIZED
	return true
}

func (chanman *myChannelManager) Close() bool {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()
	if chanman.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
		return false
	}
	close(chanman.reqCh)
	close(chanman.respCh)
	close(chanman.itemCh)
	close(chanman.errCh)
	chanman.status = CHANNEL_MANAGER_STATUS_CLOSED
	return true
}

func (chanman *myChannelManager) ReqChan() (chan base.Request, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()

	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.reqCh, nil
}

func (chanman *myChannelManager) RespChan() (chan base.Response, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()

	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.respCh, nil
}

func (chanman *myChannelManager) ItemChan() (chan base.Item, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()

	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.itemCh, nil
}

func (chanman *myChannelManager) ErrChan() (chan error, error) {
	chanman.rwmutex.Lock()
	defer chanman.rwmutex.Unlock()

	if err := chanman.checkStatus(); err != nil {
		return nil, err
	}
	return chanman.errCh, nil
}

//管理器状态
func (chanman *myChannelManager) Status() ChannelManagerStatus {
	return chanman.status
}

//摘要信息
func (chanman *myChannelManager) Summary() string {
	return fmt.Sprintf(chanmanSummaryTemplate, statusNameMap[chanman.status],
		len(chanman.reqCh), cap(chanman.reqCh),
		len(chanman.respCh), cap(chanman.respCh),
		len(chanman.itemCh), cap(chanman.itemCh),
		len(chanman.errCh), cap(chanman.errCh),
	)
}

func (chanman *myChannelManager) checkStatus() error {
	if chanman.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	statusName, ok := statusNameMap[chanman.status]
	if !ok {
		statusName = fmt.Sprintf("%d", chanman.status)
	}

	return fmt.Errorf("The undesirable status of channel managet:%s\n", statusName)
}
