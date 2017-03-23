package base

import (
	"errors"
	"fmt"
)

type ChannelArgs struct {
	reqChanLen   uint   //请求通道长度
	respChanLen  uint   //响应通道长度
	itemChanLen  uint   //条目通道长度
	errorChanLen uint   //错误通道长度
	description  string //描述
}

func NewChannelArgs(reqCL uint, respCL uint, itemCL uint, errorCL uint) ChannelArgs {
	return ChannelArgs{reqChanLen: reqCL,
		respChanLen:  respCL,
		itemChanLen:  itemCL,
		errorChanLen: errorCL,
		description:  fmt.Sprintf("reqChanLen: %d\n respChanLen: %d\n itemChanLen: %d\n errorChanLen: %d\n", reqCL, respCL, itemCL, errorCL)}

}

func (args *ChannelArgs) ReqChanLen() uint {
	return args.reqChanLen
}

func (args *ChannelArgs) RespChanLen() uint {
	return args.respChanLen
}

func (args *ChannelArgs) ItemChanLen() uint {
	return args.itemChanLen
}

func (args *ChannelArgs) ErrorChanLen() uint {
	return args.errorChanLen
}

func (args *ChannelArgs) Check() error {
	if args.reqChanLen == 0 {
		return errors.New("reqChanLen can not be 0!")
	}
	if args.respChanLen == 0 {
		return errors.New("respChanLen can not be 0!")
	}
	if args.itemChanLen == 0 {
		return errors.New("itemChanLen can not be 0!")
	}
	if args.errorChanLen == 0 {
		return errors.New("errotChanLen can not be 0!")
	}
	return nil
}

func (args *ChannelArgs) String() string {
	return args.description
}

type PoolBaseArgs struct {
	pageDownloaderPoolSize uint32 //网页下载器池尺寸
	analyzerPoolSize       uint32 //分析器池尺寸
	description            string //描述
}

func NewPoolBaseArgs(pdpSize uint32, apSize uint32) PoolBaseArgs {
	return PoolBaseArgs{pageDownloaderPoolSize: pdpSize,
		analyzerPoolSize: apSize,
		description:      fmt.Sprintf("pageDownloaderPoolSize: %d\n,analyzerPoolSize: %d\n", pdpSize, apSize)}
}

func (args *PoolBaseArgs) PageDownloaderPoolSize() uint32 {
	return args.pageDownloaderPoolSize
}

func (args *PoolBaseArgs) AnalyzerPoolSize() uint32 {
	return args.analyzerPoolSize
}

func (args *PoolBaseArgs) Check() error {
	if args.pageDownloaderPoolSize == 0 {
		return errors.New("pageDownloaderPoolSize can not to be 0!")
	}
	if args.analyzerPoolSize == 0 {
		return errors.New("anaylzer can not to be 0!")
	}
	return nil
}

func (args *PoolBaseArgs) String() string {
	return args.description
}
