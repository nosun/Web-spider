package Downloader

import (
	"github.com/yanchenxu/Web-spider/middleware"
	"reflect"
)

type myDownloaderPool struct {
	pool  middleware.Pool //实体池
	etype reflect.Type    //池内实体类型
}

type GenPageDownloader func() PageDownloader

func NewPageDownloaderPool(total uint32, gen GenPageDownloader) (PageDownloader, error) {
	etype := reflect.TypeOf(gen())

	genEntity := func() middleware.Entity {
		return gen()
	}

	pool, err := middleware.NewPool(total, etype, genEntity())
	if err != nil {
		return nil, err
	}
	return &myDownloaderPool{pool:pool, etype:etype}, nil
}
