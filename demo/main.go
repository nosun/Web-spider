package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"strings"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yanchenxu/Web-spider/analyzer"
	"github.com/yanchenxu/Web-spider/base"
	"github.com/yanchenxu/Web-spider/itemProcessor"
	"github.com/yanchenxu/Web-spider/scheduler"
)

func genHttpClient() *http.Client {
	return &http.Client{}
}

func getResponseParsers() []analyzer.ParseResponse {
	parsers := []analyzer.ParseResponse{
		parseForAtag,
	}
	return parsers
}
func parseForAtag(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	//todo 支持更多的HTTP响应状态
	if httpResp.StatusCode != 200 {
		err := fmt.Errorf("Unsuppored status code %d. (httpResponse=%v)", httpResp.StatusCode, httpResp)
		return nil, []error{err}
	}

	var reqUrl *url.URL = httpResp.Request.URL
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()

	dataList := make([]base.Data, 0)
	errs := make([]error, 0)

	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}

	//查找"A"标签并提取连接地址
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exits := sel.Attr("href")
		//前期过滤
		if !exits || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)

		//todo 对javascript代码的解析

		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}
			httpreq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := base.NewRequest(httpreq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			imap := make(map[string]interface{})
			imap["a.text"] = text
			imap["parent_url"] = reqUrl
			item := base.Item(imap)
			dataList = append(dataList, &item)
		}
	})
	return dataList, errs
}

func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invaild item!")
	}

	//生成结果
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)

	return result, nil
}

func getItemProcessors() []ItemProcessor.ProcessItem {
	itemProcessors := []ItemProcessor.ProcessItem{
		processItem,
	}
	return itemProcessors
}

func main() {
	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolBaseArgs := base.NewPoolBaseArgs(3, 3)
	crawlDepth := uint32(1)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "http//www.sogou.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}

	scheduler := scheduler.NewScheduler()

	//启动
	scheduler.Start(channelArgs,
		poolBaseArgs,
		crawlDepth,
		httpClientGenerator,
		respParsers,
		itemProcessors,
		firstHttpReq)

}
