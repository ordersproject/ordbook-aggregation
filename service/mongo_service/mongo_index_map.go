package mongo_service

import (
	"sync"
)

var indexItemMap *IndexItemMap

func init() {
	indexItemMap = NewIndexItemMap()
}

type IndexItemMap struct {
	IndexMap  map[string][]string
	Lock          *sync.RWMutex
}

func NewIndexItemMap() *IndexItemMap {
	indexItemMap := &IndexItemMap{
		IndexMap: make(map[string][]string),
		Lock: new(sync.RWMutex),
	}
	return indexItemMap
}

func (i IndexItemMap) Get(collectionName string) ([]string, bool)  {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	if v, ok := i.IndexMap[collectionName]; ok {
		return v, true
	}
	return nil, false
}

func (i IndexItemMap) CheckIndexName(collectionName, indexName string) bool  {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	if _, ok := i.IndexMap[collectionName]; ok {
		if i.IndexMap[collectionName] != nil {
			for _, v := range i.IndexMap[collectionName] {
				if v == indexName {
					return true
				}
			}
		}
	}
	return false
}

func (i IndexItemMap) Set(collectionName, indexName string)  {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if _, ok := i.IndexMap[collectionName]; ok {
		newIndeList := make([]string, 0)
		newIndeList = append(newIndeList, indexName)
		if i.IndexMap[collectionName] != nil {
			for _, v := range i.IndexMap[collectionName] {
				if v == indexName {
					continue
				}
				newIndeList = append(newIndeList, v)
			}
		}
		i.IndexMap[collectionName] = newIndeList
	}else {
		i.IndexMap[collectionName] = make([]string, 0)
	}
}


func (i IndexItemMap) Deleted(collectionName string)  {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	newIndexMap := make(map[string][]string)
	for m, v := range i.IndexMap {
		if m == collectionName {
			continue
		}
		newIndexMap[m] = v
	}
	i.IndexMap = nil
	i.IndexMap = newIndexMap
}
