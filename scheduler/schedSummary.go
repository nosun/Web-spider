package scheduler

import (
	"bytes"
	"fmt"
	"github.com/yanchenxu/Web-spider/base"
)

type mySchedSummary struct {
	prefix       string            //前缀
	running      uint32            //运行标记
	channelArgs  base.ChannelArgs  //通道参数容器
	poolBaseArgs base.PoolBaseArgs //池基本参数容器
	crawlDepth   uint32            //爬取最大深度

	chanmanSummary      string // 通道管理器的摘要信息
	reqCacheSummary     string // 请求缓存的摘要信息
	itemPipelineSummary string // 条目处理管道的摘要信息
	stopSignSummary     string // 停止信号的摘要信息

	dlPoolLen       uint32 // 网页下载器池的长度
	dlPoolCap       uint32 // 网页下载器池的容量
	analyzerPoolLen uint32 // 分析器池的长度
	analyzerPoolCap uint32 // 分析器池的容量

	urlCount  int    //已请求的URL的计数
	urlDetail string //已请求的URL的详细信息
}

func NewSchedSummary(sched *myScheduler, prefix string) SchedSummary {
	if sched == nil {
		return nil
	}

	urlCount := len(sched.urlMap)
	var urlDetail string
	if urlCount > 0 {
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for k, _ := range sched.urlMap {
			buffer.WriteString(prefix)
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		urlDetail = buffer.String()
	} else {
		urlDetail = "\n"
	}
	return &mySchedSummary{
		prefix:       prefix,
		running:      sched.running,
		channelArgs:  sched.channelArgs,
		poolBaseArgs: sched.poolBaseArgs,
		crawlDepth:   sched.crawlDepth,

		chanmanSummary:      sched.chanman.Summary(),
		reqCacheSummary:     sched.reqCache.summary(),
		stopSignSummary:     sched.stopSign.Summary(),
		itemPipelineSummary: sched.itemPipeline.Summary(),

		dlPoolLen:       sched.dlpool.Used(),
		dlPoolCap:       sched.dlpool.Total(),
		analyzerPoolLen: sched.analyzerPool.Used(),
		analyzerPoolCap: sched.analyzerPool.Total(),

		urlCount:  urlCount,
		urlDetail: urlDetail,
	}
}

func (ss *mySchedSummary) String() string {
	return ss.getSummary(false)
}

func (ss *mySchedSummary) Detail() string {
	return ss.getSummary(true)
}

func (ss *mySchedSummary) Same(other SchedSummary) bool {
	if other == nil {
		return false
	}
	otherSs, ok := interface{}(other).(*mySchedSummary)
	if !ok {
		return false
	}
	if ss.running != otherSs.running ||
		ss.channelArgs != otherSs.channelArgs ||
		ss.poolBaseArgs != otherSs.poolBaseArgs ||
		ss.crawlDepth != otherSs.crawlDepth ||

		ss.dlPoolLen != otherSs.dlPoolLen ||
		ss.dlPoolCap != otherSs.dlPoolCap ||
		ss.analyzerPoolLen != otherSs.analyzerPoolLen ||
		ss.analyzerPoolCap != otherSs.analyzerPoolCap ||

		ss.urlCount != otherSs.urlCount ||
		ss.stopSignSummary != otherSs.stopSignSummary ||
		ss.reqCacheSummary != otherSs.reqCacheSummary ||
		ss.itemPipelineSummary != otherSs.itemPipelineSummary ||
		ss.chanmanSummary != otherSs.chanmanSummary {
		return false
	} else {
		return true
	}

}

//获取再要信息
func (ss *mySchedSummary) getSummary(detail bool) string {
	prefix := ss.prefix
	template := prefix + "Running: %v \n" +
		prefix + "poolBaseArgs: %d \n" +
		prefix + "channelArgs: %d \n" +
		prefix + "Crawl depth: %d \n" +
		prefix + "Channels manager: %s \n" +
		prefix + "Request cache: %s\n" +
		prefix + "Downloader pool: %d/%d\n" +
		prefix + "Analyzer pool: %d/%d\n" +
		prefix + "Item pipeline: %s\n" +
		prefix + "Urls(%d): %s" +
		prefix + "Stop sign: %s\n"

	return fmt.Sprintf(template,
		func() bool {
			return ss.running == 1
		}(), ss.poolBaseArgs,
		ss.channelArgs,
		ss.crawlDepth,
		ss.chanmanSummary,
		ss.reqCacheSummary,
		ss.dlPoolLen,
		ss.dlPoolCap,
		ss.itemPipelineSummary,
		ss.urlCount,
		func() string {
			if detail {
				return ss.urlDetail
			} else {
				return "<concealed>\n"
			}
		}(),
		ss.stopSignSummary)
}
