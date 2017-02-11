package ItemProcessor

import "github.com/yanchenxu/Web-spider/base"

type ItemPipeline interface {
	//发送条目
	Send(item base.Item) []error
	//快速失败
	FailFast() bool
	//设置是否快速失败
	SetFailFast(failFast bool)
	//获得已发送、以接受和以处理的条目的数量
	Count() []uint64
	//获取正在被处理的条目的数量
	ProcessingNumber() uint64
	//获取摘要星系
	Summary() string
}
