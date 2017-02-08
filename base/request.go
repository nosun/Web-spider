package base

import "net/http"

//请求
type Request struct {
	httpReq *http.Request //HTTP 请求的指针值
	depth   uint32          //请求的深度
}

func NewRequest(httpReq *http.Request, depth uint) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

//获取HTTP请求
func (req *Request) HttpReq() *http.Request {
	return req.httpReq
}

//获取深度值
func (req *Request) Depth() uint {
	return req.depth
}

//数据是否有效
func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}
