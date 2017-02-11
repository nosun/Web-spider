package Downloader

import (
	"fmt"
	"github.com/yanchenxu/Web-spider/middleware"
	"reflect"
)

type myDownloaderPool struct {
	pool  middleware.Pool //实体池
	etype reflect.Type    //池内实体类型
}

type GenPageDownloader func() PageDownloader

func NewPageDownloaderPool(total uint32, gen GenPageDownloader) (PageDownloaderPool, error) {
	etype := reflect.TypeOf(gen())

	genEntity := func() middleware.Entity {
		return gen()
	}

	pool, err := middleware.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myDownloaderPool{pool: pool, etype: etype}
	return dlpool, nil
}

func (dlpool *myDownloaderPool) Take() (PageDownloader, error) {
	entity, err := dlpool.pool.Take()
	if err != nil {
		return nil, err
	}

	dl, ok := entity.(PageDownloader)
	if !ok {
		panic(fmt.Errorf("The type of entity is NOT %s!\n", dlpool.etype))
	}
	return dl, nil
}

func (dlpool *myDownloaderPool) Return(dl PageDownloader) error {
	return dlpool.pool.Return(dl)
}

func (dlpool *myDownloaderPool) Total() uint32 {
	return dlpool.pool.Total()
}

func (dlpool *myDownloaderPool) Used() uint32 {
	return dlpool.pool.Used()
}
