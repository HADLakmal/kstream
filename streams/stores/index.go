package stores

import (
	"fmt"
	"github.com/gmbyapa/kstream/pkg/errors"
	"sync"
)

type KeyMapper func(key, val interface{}) (idx string)

var UnknownIndex = errors.New(`index does not exist`)

type index struct {
	indexes map[interface{}]map[interface{}]struct{} // indexKey:recordKey:bool
	mapper  func(key, val interface{}) (idx interface{})
	mu      *sync.Mutex
	name    string
}

func NewIndex(name string, mapper func(key, val interface{}) (idx interface{})) Index {
	return &index{
		indexes: make(map[interface{}]map[interface{}]struct{}),
		mapper:  mapper,
		mu:      new(sync.Mutex),
		name:    name,
	}
}

func (s *index) String() string {
	return s.name
}

func (s *index) Write(key, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	hashKey := s.mapper(key, value)
	_, ok := s.indexes[hashKey]
	if !ok {
		s.indexes[hashKey] = make(map[interface{}]struct{})
	}
	s.indexes[hashKey][key] = struct{}{}

	return nil
}

func (s *index) ValueIndexed(index, value interface{}) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.indexes[index]
	if !ok {
		return false, nil
	}

	_, ok = s.indexes[index][value]
	return ok, nil
}

func (s *index) Hash(key, val interface{}) (hash interface{}) {
	return s.mapper(key, val)
}

func (s *index) Delete(key, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	hashKey := s.mapper(key, value)
	if _, ok := s.indexes[hashKey]; !ok {
		return fmt.Errorf(`hashKey [%s] does not exist for [%s]`, hashKey, s.name)
	}

	delete(s.indexes[hashKey], key)
	return nil
}

func (s *index) Keys() []interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []interface{}

	for key := range s.indexes {
		keys = append(keys, key)
	}

	return keys
}

func (s *index) Values() map[interface{}][]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	values := make(map[interface{}][]interface{})

	for idx, keys := range s.indexes {
		for key := range keys {
			values[idx] = append(values[idx], key)
		}
	}

	return values
}

func (s *index) Read(key interface{}) ([]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var indexes []interface{}
	idx, ok := s.indexes[key]
	if !ok {
		return nil, UnknownIndex
	}

	for k := range idx {
		indexes = append(indexes, k)
	}

	return indexes, nil
}
