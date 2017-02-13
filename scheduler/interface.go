package scheduler

import (
	"github.com/yanchenxu/Web-spider/itemProcessor"
	"github.com/yanchenxu/Web-spider/analyzer"
	"net/http"
)

//调度器的接口
type scheduler interface {
	//启动调度器
	Start(channelLen uint, //指定数据传输通道长度
	poolSize uint32, //设定池的容量
	crawlDepth uint32, //爬取深度
	httpClientGenerator GenHttpClient, //生成行的HTTP客户端
	respParsers []analyzer.ParseResponse, //解析HTTP响应
	itemProcessors []ItemProcessor.ProcessItem, //条目处理序列
	firstHttpReq *http.Request) (err error) //首次请求
	//终止
	Stop() bool
	//半段调度器是否在运行
	Running() bool
	//错误通道
	ErrorChan() <-chan error
	//判断是否空闲
	Idle() bool
	//后去摘要信息
	Summary(prefix string) SchedSummary
}

//调度器摘要信息借口
type SchedSummary interface {
	String() string               //一般
	Detail() string               //详细
	Same(other SchedSummary) bool //与另一份是否相同
}
