package Downloader

import (
	"github.com/yanchenxu/Web-spider/base"
	"net/http"
	"github.com/yanchenxu/Web-spider/middleware"
)

var downloaderIDGenertor middleware.IDGenertor = middleware.NewIDGenertor()

type myDownloader struct {
	httpClient http.Client
	id         uint32
}

func NewDownloder(client http.Client) PageDownloader {
	if client == nil {
		client = &http.Client{}
	}
	id := genDownloaderID()
	return &myDownloader{httpClient:client, id:id}
}

func genDownloaderID() uint32 {
	return downloaderIDGenertor.GetUint32()
}

func (dl *myDownloader)ID() uint32 {
	return dl.id
}

func (dl *myDownloader)Download(req base.Request) (*base.Response, error) {
	httpReq := req.HttpReq()

	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	return base.NewResponse(httpResp, req.Depth())
}