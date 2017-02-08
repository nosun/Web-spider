package middleware

import (
	"reflect"
	"fmt"
	"errors"
	"sync"
)

type Entity interface {
	ID() uint32
}

type myPool struct {
	total       uint32
	genEntity   func() Entity   //实体生成函数
	container   chan Entity     //实体容器
	etype       reflect.Type    //实体类型
	idContainer map[uint32]bool //实体
	mutex       sync.Mutex      //针对实体ID容器操作的互斥所
}

func NewPool(total uint32, entityType reflect.Type, genEntity func() Entity) (Pool, error) {
	if total == 0 {
		return nil, fmt.Errorf("The pool can not be initialized!(total=%d)\n", total)
	}
	container := make(chan Entity, int(total))
	idContainer := make(map[uint32]bool)
	for i := 0; i < int(total); i++ {
		newEntity := genEntity()
		if entityType != reflect.TypeOf(newEntity) {
			return nil, fmt.Errorf("The type of result of function genEntity() is NOT %s\n", entityType)
		}

		container <- newEntity
		idContainer[newEntity.ID()] = true
	}

	return &myPool{total:total, etype:entityType, genEntity:genEntity, container:container, idContainer:idContainer}, nil
}

func (mp *myPool)Take() (Entity, error) {
	entity, ok := <-mp.container; if !ok {
		return nil, errors.New("The inner container is invalid!")
	}
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.idContainer[entity.ID()] = false
	return entity, nil
}

func (mp *myPool)Return(entity Entity) error {
	if entity == nil {
		return errors.New("The returning entity is invalid")
	}
	if mp.etype != reflect.TypeOf(entity) {
		return fmt.Errorf("The type of returning entity is NOT %s\n", mp.etype)
	}
	result := mp.compareAndSetForIdContainer(entity.ID(), true)
	if result == 1 {
		mp.container <- entity
		return nil
	} else if result == -1 {
		return fmt.Errorf("The entity (id=%d) is alaready in the pool\n", entity.ID())
	} else {
		return fmt.Errorf("The entity(id=%d) is illegal!\n", entity.ID())
	}
}

func (mp*myPool)Total() uint32 {
	return mp.total
}

func (mp*myPool)Used() uint32 {
	return uint32(len(mp.idContainer))
}

//比较并设置实体ID容器中与给定实体ID对应的兼职对的元素值
//结果值： -1 表示键值对不存在
//	   0 表示操作失败
//	   1 表示操作成功
func (mp *myPool)compareAndSetForIdContainer(entityID uint32, newValue bool) int8 {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	v, ok := mp.idContainer[entityID]
	if !ok {
		return 0
	}
	if v {
		return -1
	}
	mp.idContainer[entityID] = newValue
	return 1
}

