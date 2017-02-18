package main

import (
	"github.com/yanchenxu/Web-spider/base"
	"net/http"
	"log"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"io"
	"github.com/yanchenxu/Web-spider/analyzer"
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
func parseForAtag(httpResp*http.Response, respDepyh uint32) ([]base.Data, []error) {
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
	doc.Find("a").Each(func())

}

func main() {
	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolBaseArgs := base.NewPoolBaseArgs(3, 3)
	crawlDepth := uint32(1)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	startUrl := "http//www.sogou.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
}
