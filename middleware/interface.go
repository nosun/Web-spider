package middleware

import (
	"github.com/yanchenxu/Web-spider/base"
)

//id生成器的借口类型
type IDGenertor interface {
	GetUint32() uint32
}

//通道管理器接口
type ChannelManager interface {
	//初始化
	Init(channelLen uint, rest bool) bool
	//关闭管道
	Close() bool
	//获取请求传输通道
	ReqChan() (chan base.Request, error)
	//获取响应
	RespChan() (chan base.Response, error)
	//条目
	ItemChan() (chan base.Item, error)
	//错误
	ErrChan() (chan error, error)
	//通道长度
	ChannelLen() uint
	//管理器状态
	Status() ChannelManagerStatus
	//摘要信息
	Summary() string
}

//实体池 todo 泛类型？
type Pool interface {
	Take() (Entity, error)
	Return(entity Entity) error
	Total() uint32
	Used() uint32
}

//停止信号
type StopSign interface {
	//发出停止型号，如果已经发出，返回false
	Sign() bool
	//判断停止型号是否发出
	Signed() bool
	//重置
	Reset()
	//处理
	Deal(code string)
	//获得处理计数
	DealCount(code string) uint32
	//总计数
	DealTotal() uint32
	//摘要
	Summary() string
}
