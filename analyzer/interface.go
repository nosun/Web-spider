package analyzer

import "github.com/yanchenxu/Web-spider/base"

type Analyzer interface {
	Id() uint32
	Anzlyze(respParsers []ParseResponse, resp base.Response) ([]base.Data, []error)
}

type AnalyzerPool interface {
	Take() (Analyzer, error)//从池中取出一个分析器
	Return(ananlyzer Analyzer) error //把一个分析器器归还给池
	Total() uint32 //池的总容量
	Used() uint32  //获得正在被使用的分析器的数量
}