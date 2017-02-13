package ItemProcessor

import (
	"errors"
	"fmt"
	"github.com/yanchenxu/Web-spider/base"
	"sync/atomic"
)

//被用来处理条目的函数类型
type ProcessItem func(item base.Item) (result base.Item, err error)

type myItemPipeline struct {
	itemProcessors []ProcessItem //条目处理器列表
	failFast       bool          //表示是否需要快速失败的标志
	sent           uint64        //已被发送的条目的数量
	accepted       uint64        //已被接受的条目的数量
	processed      uint64        //已被处理的条目的数量
	processing     uint64        //正在被处理的条目的数量
}

func NewItemPipeline(itemProcessors []ProcessItem) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New("Invalid item processor list!"))
	}
	innerItemProcessors := make([]ProcessItem, 0)
	for i, ip := range itemProcessors {
		if ip == nil {
			panic(fmt.Errorf("Invalid item processor【%d】!\n", i))
		}
		innerItemProcessors = append(innerItemProcessors, ip)
	}
	return &myItemPipeline{itemProcessors: innerItemProcessors}
}

//发送条目
func (ip *myItemPipeline) Send(item base.Item) []error {
	atomic.AddUint64(&ip.processing, 1)
	defer atomic.AddUint64(&ip.processing, ^uint64(0))
	atomic.AddUint64(&ip.sent, 1)

	errs := make([]error, 0)
	if item == nil {
		errs = append(errs, errors.New("The item is invalid!"))
	}
	atomic.AddUint64(&ip.accepted, 1)

	var currentItem base.Item = item
	for _, itemProcessor := range ip.itemProcessors {
		processedItem, err := itemProcessor(currentItem)
		if err != nil {
			errs = append(errs, err)
			if ip.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	atomic.AddUint64(&ip.processed, 1)
	return errs
}

//快速失败
func (ip *myItemPipeline) FailFast() bool {
	return ip.failFast
}

//设置是否快速失败
func (ip *myItemPipeline) SetFailFast(failFast bool) {
	ip.failFast = failFast
}

//获得已发送、以接受和以处理的条目的数量
func (ip *myItemPipeline) Count() []uint64 {
	counts := make([]uint64, 3)
	counts[0] = atomic.LoadUint64(&ip.sent)
	counts[1] = atomic.LoadUint64(&ip.accepted)
	counts[2] = atomic.LoadUint64(&ip.processed)
	return counts
}

//获取正在被处理的条目的数量
func (ip *myItemPipeline) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&ip.processing)
}

var summaryTemplate = "failFast :%v,processorNumber:%d," +
	"set: %d,accepted: %d,processed: %d,processing: %d"

//获取摘要星系
func (ip *myItemPipeline) Summary() string {
	counts := ip.Count()
	summary := fmt.Sprintf(summaryTemplate, ip.failFast, len(ip.itemProcessors), counts[0], counts[1], counts[2], ip.ProcessingNumber())
	return summary
}
