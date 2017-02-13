package base

import "net/http"

//响应
type Response struct {
	httpResp *http.Response
	depth    uint32
}

func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

//获取HTTP请求
func (resp *Response) HttpReq() *http.Response {
	return resp.httpResp
}

//获取深度值
func (resp *Response) Depth() uint32 {
	return resp.depth
}

//数据是否有效
func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}
