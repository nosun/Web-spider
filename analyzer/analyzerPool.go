package analyzer

import (
	"fmt"
	"github.com/yanchenxu/Web-spider/middleware"
	"reflect"
)

type myAnalyzerPool struct {
	pool  middleware.Pool
	etype reflect.Type
}

type genAnalyzer func() Analyzer

func NewAnalyzerPool(total uint32, gen genAnalyzer) (AnalyzerPool, error) {
	etype := reflect.TypeOf(gen())

	genEntity := func() middleware.Entity {
		return gen()
	}
	pool, err := middleware.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}

	return &myAnalyzerPool{pool: pool, etype: etype}, nil
}

func (aPool *myAnalyzerPool) Take() (Analyzer, error) {
	entity, err := aPool.pool.Take()
	if err != nil {
		return nil, err
	}

	analyzer, ok := entity.(Analyzer)
	if !ok {
		panic(fmt.Errorf("The type of entity is NOT %s!\n", aPool.etype))
	}
	return analyzer, nil
}

func (aPool *myAnalyzerPool) Return(ananlyzer Analyzer) error {
	return aPool.pool.Return(ananlyzer)
}

func (aPool *myAnalyzerPool) Total() uint32 {
	return aPool.pool.Total()
}

func (aPool *myAnalyzerPool) Used() uint32 {
	return aPool.pool.Used()
}
