package base

import (
	"bytes"
	"fmt"
)

type ErrorType string

//错误类型常量
const (
	DOWNLOADER_ERROR     ErrorType = "Downloader Error"
	ANAYZER_ERROR        ErrorType = "Analyzer Error"
	ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

type myCrawlerError struct {
	errType    ErrorType //错误类型
	errMsg     string    //错误提示信息
	fullErrMsg string    //完整的错误提示信息
}

func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{errType: errType, errMsg: errMsg}
}

func (ce *myCrawlerError) Type() ErrorType {
	return ce.errType
}

func (ce *myCrawlerError) Error() string {
	if ce.fullErrMsg == "" {
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}

//生成错误信息，并给相应的字段复制
func (ce *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error: ")
	if ce.errType != "" {
		buffer.WriteString(string(ce.errType) + ": ")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrMsg = fmt.Sprintf("%s\n", buffer.String())
}
