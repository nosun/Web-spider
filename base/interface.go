package base

type Data interface {
	Valid() bool
}

type CrawlerError interface {
	Type() ErrorType //获得错误类型
	Error() string   //获得错误提示信息
}