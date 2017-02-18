package scheduler

import (
	"errors"
	"fmt"
	"github.com/yanchenxu/Web-spider/analyzer"
	"github.com/yanchenxu/Web-spider/base"
	"github.com/yanchenxu/Web-spider/downloader"
	"github.com/yanchenxu/Web-spider/itemProcessor"
	"github.com/yanchenxu/Web-spider/middleware"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type GenHttpClient func() *http.Client

type myScheduler struct {
	channelArgs   base.ChannelArgs  //通道参数容器
	poolBaseArgs  base.PoolBaseArgs //池基本参数-容器
	crawlDepth    uint32            //爬取最大深度
	primaryDomain string            //主域名

	chanman      middleware.ChannelManager     //通道管理器
	stopSign     middleware.StopSign           //停止信号
	dlpool       Downloader.PageDownloaderPool //网页下载池
	analyzerPool analyzer.AnalyzerPool         //分析池
	itemPipeline ItemProcessor.ItemPipeline    //条目处理管道

	running uint32 //运行标记，0运行，1已运行，2停止

	reqCache requestCache //请求缓存

	urlMap map[string]bool //已请求的URL的字典
}

const (
	DOWNLOADER_CODE   = "downloader"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "item_PIPELINE"
	SCHEDULER_CODE    = "scheduler"
)

func NewScheduler() scheduler {
	return &myScheduler{}
}

func (sched *myScheduler) Start(channelArgs base.ChannelArgs, //代表通道参数容器
	poolBaseArgs base.PoolBaseArgs, //代表池基本参数的容器
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

	if atomic.LoadUint32(&sched.running) == 1 {
		return errors.New("The scheduler has been started!\n")
	}

	atomic.StoreUint32(&sched.running, 1)

	if err := channelArgs.Check(); err != nil {
		return err
	}
	sched.channelArgs = channelArgs

	if err := poolBaseArgs.Check(); err != nil {
		return err
	}
	sched.poolBaseArgs = poolBaseArgs
	sched.crawlDepth = crawlDepth

	sched.chanman = generateChannelManager(sched.channelArgs)

	if httpClientGenerator == nil {
		return errors.New("The HTTP client generator list is invalid!")
	}

	dlpool, err := generatePageDownloaderPool(sched.poolBaseArgs, httpClientGenerator)
	if err != nil {
		return fmt.Errorf("Occur error when get page downloader pool: %s\n", err)
	}
	sched.dlpool = dlpool

	analyzerPool, err := generateAnalyzerPool(sched.poolBaseArgs)
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

//startDownloading
func (sched *myScheduler) startDownloading() {
	go func() {
		for {
			req, ok := <-sched.getReqChan()
			if !ok {
				break
			}
			go sched.download(req)
		}

	}()
}

func (sched *myScheduler) download(req base.Request) {
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

//activateAnalyzers
func (sched *myScheduler) activateAnalyzers(respParsers []analyzer.ParseResponse) {
	go func() {
		for {
			resp, ok := <-sched.getRespChan()
			if !ok {
				break
			}
			go sched.analyze(respParsers, resp)
		}
	}()
}

func (sched *myScheduler) analyze(respParsers []analyzer.ParseResponse, resp base.Response) {
	defer func() {
		if p := recover(); p != nil {
			log.Fatalf("Fatal Analysis Error: %s\n", p)
		}
	}()

	analyzer, err := sched.analyzerPool.Take()
	if err != nil {
		sched.sendError(fmt.Errorf("Analyuzer pool error: %s\n", err), SCHEDULER_CODE)
		return
	}

	defer func() {
		err := sched.analyzerPool.Return(analyzer)
		if err != nil {
			sched.sendError(fmt.Errorf("Analyuzer pool error: %s\n", err), SCHEDULER_CODE)
		}
	}()

	code := generateCode(ANALYZER_CODE, analyzer.ID())
	datalist, errs := analyzer.Analyzer(respParsers, resp)

	if datalist != nil {
		for _, data := range datalist {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *base.Request:
				sched.saveReqToCache(*d, code)
			case *base.Item:
				sched.sendItem(*d, code)
			default:
				sched.sendError(fmt.Errorf("Unsupported data type '%T'!(value=%v)\n", d, d), code)
			}
		}
	}

	if errs != nil {
		for _, err := range errs {
			if err != nil {
				sched.sendError(err, code)
			}
		}
	}

}

func (sched *myScheduler) saveReqToCache(req base.Request, code string) bool {
	httpReq := req.HttpReq()
	if httpReq == nil {
		log.Fatalln("WARN:Ignore the request! It's HTTP request is Iinvalid!")
		return false
	}
	reqUrl := httpReq.URL
	if reqUrl == nil {
		log.Fatalln("WARN:Ignore the request! It's url is invalid!")
		return false
	}
	if strings.ToLower(httpReq.URL.Scheme) != "http" {
		log.Fatalf("WARN:Iggnore the request! It's url scheme '%s',but should be 'http'!\n", reqUrl.Scheme)
		return false
	}

	if _, ok := sched.urlMap[reqUrl.String()]; ok {
		log.Fatalf("WARN:Ignore teh request! It's url is repeated.(requestUrl=%s)\n", reqUrl)
		return false
	}

	if pd, _ := getPrimaryDomain(httpReq.Host); pd != sched.primaryDomain {
		log.Fatalf("WARN:Ignore the request It's host '%s' not in primary domain '%s',(requestUrl=%s)\n", httpReq.Host, sched.primaryDomain, reqUrl)
		return false
	}

	if req.Depth() > sched.crawlDepth {
		log.Fatalf("WARN:Ignore the request! It's depth %d greater than %d.(requestUrl=%s)\n", req.Depth(), sched.crawlDepth, reqUrl)
		return false
	}

	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}

	sched.reqCache.put(&req)

	sched.urlMap[reqUrl.String()] = true
	return true
}

//openItemPipeline
func (sched *myScheduler) openItemPipeline() {
	go func() {
		sched.itemPipeline.SetFailFast(true)
		code := ITEMPIPELINE_CODE
		for item := range sched.getItemChan() {
			go func(item base.Item) {
				defer func() {
					if p := recover(); p != nil {
						log.Fatal(fmt.Errorf("Fatal Item Processing Error: %s\n", p))
					}
				}()
				errs := sched.itemPipeline.Send(item)
				if errs != nil {
					for _, err := range errs {
						sched.sendError(err, code)
					}
				}
			}(item)
		}
	}()
}

//schedule
func (sched *myScheduler) schedule(interval time.Duration) {
	go func() {
		for {
			if sched.stopSign.Signed() {
				sched.stopSign.Deal(SCHEDULER_CODE)
				return
			}
			remainder := cap(sched.getReqChan()) - len(sched.getReqChan())
			var temp *base.Request
			for remainder > 0 {
				temp = sched.reqCache.get()
				if temp == nil {
					break
				}
				if sched.stopSign.Signed() {
					sched.stopSign.Deal(SCHEDULER_CODE)
					return
				}
				sched.getReqChan() <- *temp
				remainder--
			}
			time.Sleep(interval)
		}
	}()
}

func (sched *myScheduler) Stop() bool {
	if atomic.LoadUint32(&sched.running) != 1 {
		return false
	}
	sched.stopSign.Sign()
	sched.chanman.Close()
	sched.reqCache.close()
	atomic.StoreUint32(&sched.running, 2)
	return true
}

func (sched *myScheduler) Running() bool {
	return atomic.LoadUint32(&sched.running) == 1
}

func (sched *myScheduler) ErrorChan() <-chan error {
	if sched.chanman.Status() != middleware.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return sched.getErrorChan()
}

func (sched *myScheduler) Idle() bool {
	idleDlPool := sched.dlpool.Used() == 0
	idleAnalyzerPool := sched.analyzerPool.Used() == 0
	idleItemPipeline := sched.itemPipeline.ProcessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleItemPipeline {
		return true
	}
	return false
}

func (sched *myScheduler) Summary(prefix string) SchedSummary {
	return NewSchedSummary(sched, prefix)
}

func (sched *myScheduler) sendResp(resp base.Response, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getRespChan() <- resp

	return true
}

func (sched *myScheduler) sendItem(item base.Item, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}

	sched.getItemChan() <- item
	return true
}

func (sched *myScheduler) sendError(err error, code string) bool {
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

func (sched *myScheduler) getReqChan() chan base.Request {
	reqChan, err := sched.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}

func (sched *myScheduler) getRespChan() chan base.Response {
	respChan, err := sched.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

func (sched *myScheduler) getErrorChan() chan error {
	errChan, err := sched.chanman.ErrChan()
	if err != nil {
		panic(err)
	}
	return errChan
}

func (sched *myScheduler) getItemChan() chan base.Item {
	itemChan, err := sched.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}
