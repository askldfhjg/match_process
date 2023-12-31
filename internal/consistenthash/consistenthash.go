/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package consistenthash provides an implementation of a ring hash.
package consistenthash

import (
	"context"
	"hash/crc32"
	"match_process/internal/db"
	"sort"
	"strconv"

	"github.com/micro/micro/v3/service/logger"
)

type Hash func(data []byte) uint32

type HashRing struct {
	hash     Hash
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *HashRing {
	m := &HashRing{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty returns true if there are no items available.
func (m *HashRing) IsEmpty() bool {
	return len(m.keys) == 0
}

func (r *HashRing) Reset(nodes ...string) {
	// 先清空
	r.keys = nil
	r.hashMap = map[int]string{}
	// 再重置
	r.Add(nodes...)
}

// Add adds some keys to the hash.
func (m *HashRing) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *HashRing) Remove(key string) {
	remove := make(map[int]int, m.replicas)
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		delete(m.hashMap, hash)
		remove[hash] = 1
	}
	ret := make([]int, 0, len(m.keys))
	for _, item := range m.keys {
		if _, ok := remove[item]; !ok {
			ret = append(ret, item)
		}
	}
	m.keys = ret
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *HashRing) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.keys) {
		idx = 0
	}
	retKey := m.hashMap[m.keys[idx]]
	ret, err := db.Default.SetEvalUrl(context.Background(), key, retKey)
	if err == nil {
		return ret
	} else {
		logger.Errorf("ring get error %s", err.Error())
		return m.hashMap[m.keys[0]]
	}
}
