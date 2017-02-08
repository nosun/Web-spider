package analyzer

import (
	"net/http"
	"github.com/yanchenxu/Web-spider/base"
)
//被用于解析HTTP响应的函数类型
type ParseResponse func(httpResp *http.Response, resoDeth uint32) ([]base.Data, []error)
