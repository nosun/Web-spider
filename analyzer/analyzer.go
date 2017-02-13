package analyzer

import (
	"github.com/yanchenxu/Web-spider/base"
	"github.com/yanchenxu/Web-spider/middleware"
	"net/http"
	"errors"
	"net/url"
	"fmt"
)

//被用于解析HTTP响应的函数类型
type ParseResponse func(httpResp *http.Response, respDeth uint32) ([]base.Data, []error)

var genAnalyzerIDGenertor middleware.IDGenertor = middleware.NewIDGenertor()

type myAnalyzer struct {
	id uint32
}

func NewAnalyzer() Analyzer {
	return &myAnalyzer{id:genAnalyzerIDGenertor.GetUint32()}
}

func (analyzer *myAnalyzer)ID() uint32 {
	return analyzer.id
}

func (Analyzer *myAnalyzer)Analyzer(respParsers []ParseResponse, resp base.Response) (dataList []base.Data, errorList[]error) {
	if respParsers == nil {
		return nil, []error{errors.New("The response parser list id invalid!")}
	}

	httpResp := resp.HttpReq()
	if httpResp == nil {
		return nil, []error{errors.New("The http response is invalid")}
	}

	var reqUrl *url.URL = httpResp.Request.URL
	//todo 日志
	fmt.Printf("Parse the response (reqUrl = %s)", reqUrl)

	respDepth := resp.Depth()

	dataList = make([]base.Data, 0)
	errorList = make([]error, 0)

	for i, respParser := range respParsers {
		pDataList, pErrorList := respParser(httpResp, respDepth)
		if respParser == nil {
			errorList = append(errorList, fmt.Errorf("The document parser {%d}is invalid!", i))
		}

		if pDataList != nil {
			for _, pData := range pDataList {
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}
		if pErrorList != nil {
			for _, pError := range errorList {
				errorList = appendErrorList(errorList, pError)
			}
		}

	}

	return dataList, errorList
}

func appendDataList(dataList []base.Data, data base.Data, respDeth uint32) []base.Data {
	if data == nil {
		return dataList
	}

	req, ok := data.(*base.Request)
	if !ok {
		return append(dataList, data)
	}
	newDepth := respDeth + 1
	if req.Depth() != newDepth {
		req = base.NewRequest(req.HttpReq(), newDepth)
	}
	return append(dataList, req)
}

func appendErrorList(errorList []error, err error) []error {
	if err == nil {
		return errorList
	}
	return append(errorList, err)
}