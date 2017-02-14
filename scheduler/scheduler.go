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
	"sync/atomic"
	"time"
)

type GenHttpClient func() *http.Client

type myScheduler struct {
	poolSize      uint32                        //池的尺寸
	channelLen    uint                          //通道的总长度
	crawlDepth    uint32                        //爬取最大深度
	primaryDomain string                        //主域名

	chanman       middleware.ChannelManager     //通道管理器
	stopSign      middleware.StopSign           //停止信号
	dlpool        Downloader.PageDownloaderPool //网页下载池
	analyzerPool  analyzer.AnalyzerPool         //分析池
	itemPipeline  ItemProcessor.ItemPipeline    //条目处理管道

	running       uint32                        //运行标记，0运行，1已运行，2停止

	reqCache      requestCache                  //请求缓存

	urlMap        map[string]bool               //已请求的URL的字典
}

const (
	DOWNLOADER_CODE = "downloader"
	ANALYZER_CODE = "analyzer"
	ITEMPIPELINE_CODE = "item_PIPELINE"
	SCHEDULER_CODE = "scheduler"
)

func NewScheduler() scheduler {
	return &myScheduler{}
}

func (sched *myScheduler)Start(channelLen uint, //指定数据传输通道长度
poolSize uint32, //设定池的容量
crawlDepth uint32, //爬取深度
httpClientGenerator GenHttpClient, //生成行的HTTP客户端
respParsers []analyzer.ParseResponse, //解析HTTP响应
itemProcessors []ItemProcessor.ProcessItem, //条目处理序列
firstHttpReq *http.Request) (err error) {

	defer func() {
		if p := recover(); p != nil {
			fmt.Println(fmt.Sprintf("Fatal Scheduler Error: %s\n", p))
			err = errors.New(fmt.Sprintf("Fatal Scheduler Error: %s\n", p))
		}
	}()

	if atomic.LoadUint32(sched.running) == 1 {
		return errors.New("The scheduler has been started!\n")
	}

	atomic.StoreUint32(&sched.running, 1)

	if channelLen == 0 {
		return errors.New("The channel max length (capacity) can not be 0!\n")
	}
	sched.channelLen = channelLen

	if poolSize == 0 {
		return errors.New("The pool size can not be of 0!\n")
	}
	sched.poolSize = poolSize
	sched.crawlDepth = crawlDepth

	sched.chanman = generateChannelManager(sched.channelLen)

	if httpClientGenerator == nil {
		return errors.New("The HTTP client generator list is invalid!")
	}

	dlpool, err := generatePageDownloaderPool(sched.poolSize, httpClientGenerator)
	if err != nil {
		return fmt.Errorf("Occur error when get page downloader pool: %s\n", err)
	}
	sched.dlpool = dlpool

	analyzerPool, err := generateAnalyzerPool(sched.poolSize)
	if err != nil {
		return fmt.Errorf("Occur error when get page analyzer pool: %s\n", err)
	}
	sched.analyzerPool = analyzerPool

	if itemProcessors == nil {
		return errors.New("The item processor list is invalid!")
	}
	for i, ip := range itemProcessors {
		if ip == nil {
			return fmt.Errorf("The %dth item processor is invalid!", i)
		}
	}
	sched.itemPipeline = generateItemPipeline(itemProcessors)

	if sched.stopSign == nil {
		sched.stopSign = middleware.NewStopSign()
	} else {
		sched.stopSign.Reset()
	}
	sched.urlMap = make(map[string]bool)
	sched.reqCache = newRequestCache()

	sched.startDownloading()
	sched.activateAnalyzers(respParsers)
	sched.openItemPipeline()
	sched.schedule(10 * time.Millisecond)

	if firstHttpReq == nil {
		return errors.New("The first HTTP request id invalid!")
	}
	pd, err := getPrimaryDomain(firstHttpReq.Host)
	if err != nil {
		return err
	}
	sched.primaryDomain = pd

	firstReq := base.NewRequest(firstHttpReq, 0)
	sched.reqCache.put(firstReq)

	return nil
}

func (sched *myScheduler)startDownloading() {
	go func() {
		req, ok := <-sched.getReqChan()
		if !ok {
			break
		}
		go sched.download(req)
	}()
}

func (sched *myScheduler)download(req base.Request) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("FATAL DOwnload Error: %s\n", p)
		}
	}()

	downloader, err := sched.dlpool.Take()
	if err != nil {
		sched.sendError(fmt.Errorf("Downloader pool error: %s", err), SCHEDULER_CODE)
		return
	}

	defer func() {
		err := sched.dlpool.Return(downloader)
		if err != nil {
			sched.sendError(fmt.Errorf("Downloader pool error: %s", err), SCHEDULER_CODE)
		}
	}()

	code := generateCode(DOWNLOADER_CODE, downloader.ID())
	respp, err := downloader.Download(req)
	if respp != nil {
		sched.sendResp(*respp, code)
	}

	if err != nil {
		sched.sendError(err, code)
	}

}

func (sched *myScheduler)sendResp(resp base.Request, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getRespChan() <- resp

	return true
}

func (sched *myScheduler)sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType base.ErrorType
	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = base.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = base.ANAYZER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = base.ITEM_PROCESSOR_ERROR
	}
	cError := base.NewCrawlerError(errorType, err.Error())

	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	go func() {
		sched.getErrorChan() <- cError
	}()

	return true
}

func (sched *myScheduler)getReqChan() chan base.Request {
	reqChan, err := sched.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}

func (sched *myScheduler)getRespChan() chan base.Response {
	respChan, err := sched.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

func (sched *myScheduler)getErrorChan() chan base.CrawlerError {
	errChan, err := sched.chanman.ErrChan()
	if err != nil {
		panic(err)
	}
	return errChan
}

func (sched *myScheduler)getItemChan() chan base.Item {
	itemChan, err := sched.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}