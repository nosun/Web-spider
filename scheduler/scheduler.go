package scheduler

import (
	"net/http"
	"github.com/yanchenxu/Web-spider/middleware"
	"github.com/yanchenxu/Web-spider/itemProcessor"
	"github.com/yanchenxu/Web-spider/base"
	"github.com/yanchenxu/Web-spider/downloader"
	"github.com/yanchenxu/Web-spider/analyzer"
	"fmt"
	"errors"
)

type GenHttpClient func() *http.Client

type myScheduler struct {
	poolSize uint32 //池的尺寸
	channelLen uint//通道的总长度
	crawlDepth uint32//爬取最大深度
	primaryDomain string //主域名

	chanman middleware.ChannelManager //通道管理器
	stopSign middleware.StopSign//停止信号
	dlpool Downloader.PageDownloaderPool//网页下载池
	analyzerPool analyzer.AnalyzerPool//分析池
	itemPipeline ItemProcessor.ItemPipeline//条目处理管道

	running uint32//运行标记，0运行，1已运行，2停止

	reqCache requestCache //请求缓存

	urlMap map[string]bool//已请求的URL的字典
}

func NewScheduler()scheduler{
	return &myScheduler{}
}

func (s *myScheduler)Start(channelLen uint, //指定数据传输通道长度
	poolSize uint32, //设定池的容量
	crawlDepth uint32, //爬取深度
	httpClientGenerator GenHttpClient, //生成行的HTTP客户端
	respParsers []analyzer.ParseResponse, //解析HTTP响应
	itemProcessors []ItemProcessor.ProcessItem, //条目处理序列
	firstHttpReq *http.Request)(err error){

	defer func(){
		if p:=recover();p!=nil{
			fmt.Println(fmt.Sprintf("Fatal Scheduler Error: %s\n",p))
			err=errors.New(fmt.Sprintf("Fatal Scheduler Error: %s\n",p))
		}
	}()

}