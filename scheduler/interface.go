package scheduler

import (
	"net/http"

	"github.com/yanchenxu/Web-spider/analyzer"
	"github.com/yanchenxu/Web-spider/base"
	"github.com/yanchenxu/Web-spider/itemProcessor"
)

//调度器的接口
type scheduler interface {
	//启动调度器
	Start(channelArgs base.ChannelArgs, //通道参数容器
		poolBaseSize base.PoolBaseArgs, //池基本参数容器
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

//请求缓存的接口类型

type requestCache interface {
	//将请求放入请求缓存
	put(req *base.Request) bool
	//从请求缓存获取最早被放入且仍在其中的请求
	get() *base.Request
	//获得请求缓存的容量
	capacity() int
	//获得请求缓存的实时长度，即其中的请求的及时数量
	length() int
	//关闭请求缓存
	close()
	//获取请求缓存的摘要信息
	summary() string
}
