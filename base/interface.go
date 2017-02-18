package base

type Data interface {
	Valid() bool
}

type CrawlerError interface {
	Type() ErrorType //获得错误类型
	Error() string   //获得错误提示信息
}

//惨数容器借口
type Args interface {
	//自检参数有效性,并在必要时返回可以说明问题的错误之
	//若结果值为nil,则说明未发现问题，否则就意味着自检为通过
	Check() error
	//获得参数容器的字符串表现形式
	String() string
}
