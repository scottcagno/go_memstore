// -----------
// snapshot.go ::: manage store disk snapshots
// -----------
// Copyright (c) 2013-Present, Scott Cagno. All rights reserved.
// This source code is governed by a BSD-style license.

package data

import (
	"encoding/gob"
	"log"
	"os"
	"runtime"
)

// save store snapshot to disk
func (self *Store) SaveSnapshot(file string) bool {
	self.mu.Lock()
	f, err := os.Create(file)
	if err != nil {
		log.Println(err)
	}
	enc := gob.NewEncoder(f)
	for _, v := range self.Itms {
		gob.Register(v)
	}
	err = enc.Encode(&self.Itms)
	if err != nil {
		log.Println(err)
	}
	f.Close()
	self.mu.Unlock()
	runtime.GC()
	return err == nil
}

// load store snapshot from disk
func (self *Store) LoadSnapshot(file string) bool {
	self.mu.Lock()
	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
	}
	dec := gob.NewDecoder(f)
	self.Itms = nil
	err = dec.Decode(&self.Itms)
	if err != nil {
		log.Println(err)
	}
	f.Close()
	self.mu.Unlock()
	runtime.GC()
	return err == nil
}
