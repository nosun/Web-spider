package Downloader

import "github.com/yanchenxu/Web-spider/base"

//网页下载器的接口类型
type PageDownloader interface {
	ID() uint32                                        //获得ID
	Download(req base.Request) (*base.Response, error) //根据请求下在网页并返回响应
}

//网页下载器的接口类型
type PageDownloaderPool interface {
	Take() (PageDownloader, error)          //从池中取出一个网页下载器
	Return(downloader PageDownloader) error //把一个网页下载器归还给池
	Total() uint32                          //池的总容量
	Used() uint32                           //获得正在被使用的网页下载器的数量
}
