// --------
// store.go ::: data store
// --------
// Copyright (c) 2013-Present, Scott Cagno. All rights reserved.
// This source code is governed by a BSD-style license.

package data

import (
	"bytes"
	"runtime"
	"sync"
	"time"
)

// expiring item
type expItem struct {
	k   string
	ttl int64
}

// store
type Store struct {
	Itms map[string][][]byte
	Exps []*expItem
	mu   sync.Mutex
}

// return store instance
func GetStore(n int64) *Store {
	st := &Store{
		Itms: make(map[string][][]byte),
		Exps: make([]*expItem, 0),
	}
	if n > 0 {
		go st.runGC(n)
	}
	return st
}

// run garbage collector
func (self *Store) runGC(n int64) {
	self.GC()
	time.AfterFunc(time.Duration(n)*time.Second, func() { self.runGC(n) })
}

// garbage collector
func (self *Store) GC() {
	self.mu.Lock()
	now := time.Now().Unix()
	if len(self.Exps) > 0 {
		for i := 0; i < len(self.Exps); i++ {
			if self.Exps[i].ttl <= now {
				delete(self.Itms, self.Exps[i].k)
				self.Exps = append(self.Exps[:i], self.Exps[i+1:]...)
				i--
			}
		}
	}
	self.mu.Unlock()
}

// check if store has item with matching key
func (self *Store) HasKey(k string) bool {
	self.mu.Lock()
	_, ok := self.Itms[k]
	self.mu.Unlock()
	return ok
}

// set or update a value
func (self *Store) Set(k string, v []byte) bool {
	self.mu.Lock()
	self.Itms[k] = [][]byte{[]byte(v)}
	_, ok := self.Itms[k]
	self.mu.Unlock()
	return ok
}

// append val/vals
func (self *Store) App(k string, v ...[]byte) bool {
	self.mu.Lock()
	self.Itms[k] = append(self.Itms[k], v...)
	_, ok := self.Itms[k]
	self.mu.Unlock()
	return ok
}

// return val/vals for item
func (self *Store) Get(k string) [][]byte {
	self.mu.Lock()
	v, _ := self.Itms[k]
	self.mu.Unlock()
	return v
}

// delete an item
func (self *Store) Del(k string) bool {
	self.mu.Lock()
	delete(self.Itms, k)
	for i := 0; i < len(self.Exps); i++ {
		if self.Exps[i].k == k {
			self.Exps = append(self.Exps[:i], self.Exps[i+1:]...)
			i--
		}
	}
	_, ok := self.Itms[k]
	self.mu.Unlock()
	return ok == false
}

// set an item to expire
func (self *Store) Exp(k string, n int64) bool {
	self.mu.Lock()
	_, ok := self.Itms[k]
	now := time.Now().Unix() + n
	if ok {
		var updated bool
		for i, itm := range self.Exps {
			if itm.k == k {
				self.Exps[i].ttl = now
				updated = true
			}
		}
		if !updated {
			self.Exps = append(self.Exps, &expItem{k, now})
		}
	}
	self.mu.Unlock()
	return ok
}

// return time to live value of itm
func (self *Store) TTL(k string) int {
	self.mu.Lock()
	var ttl int
	for i := range self.Exps {
		if self.Exps[i].k == k {
			ttl = int((self.Exps[i].ttl - time.Now().Unix()))
			break
		}
	}
	self.mu.Unlock()
	return ttl
}

// return specific val/vals in list from item
func (self *Store) GetVal(k string, i ...int) [][]byte {
	self.mu.Lock()
	itm, ok := self.Itms[k]
	var v [][]byte
	if ok {
		switch len(i) {
		case 1:
			if i[0] <= len(itm) {
				v = append(v, self.Itms[k][i[0]])
			}
			break
		case 2:
			if i[0] <= i[1] && i[1] <= len(itm) {
				v = self.Itms[k][i[0]:i[1]]
			}
			break
		default:
			v = nil
			break
		}
	}
	self.mu.Unlock()
	return v
}

// delete a specifiv val in list from item
func (self *Store) DelVal(k string, v []byte) bool {
	self.mu.Lock()
	_, ok := self.Itms[k]
	if ok {
		for i, itm := range self.Itms[k] {
			if bytes.Equal(itm, v) {
				copy(self.Itms[k][i:], self.Itms[k][i+1:])
				self.Itms[k][len(self.Itms[k])-1] = nil
				self.Itms[k] = self.Itms[k][:len(self.Itms[k])-1]
			}
		}
	}
	self.mu.Unlock()
	return ok
}

// purge store, delete everyhing
func (self *Store) Purge() bool {
	self.mu.Lock()
	for i := 0; i < len(self.Exps); i++ {
		delete(self.Itms, self.Exps[i].k)
		self.Exps = append(self.Exps[:i], self.Exps[i+1:]...)
		i--
	}
	for k, _ := range self.Itms {
		delete(self.Itms, k)
	}
	self.mu.Unlock()
	runtime.GC()
	return len(self.Itms) == 0 && len(self.Exps) == 0
}
