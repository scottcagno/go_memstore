// --------
// serve.go ::: server
// --------
// Copyright (c) 2013-Present, Scott Cagno. All rights reserved.
// This source code is governed by a BSD-style license.

package data

/*
	initConfig(fpath string) bool		// load configuration attributes


	type Args struct {
		args [][]byte
		argc int
	}

	ArgByIndex(n int) []byte			// return []byte of arg at index n
	IndexByArg(arg []byte) int			// return n of arg
	SliceNArgs(c1, c2 int) [][]byte		// return slice of args from c1, to c2

	xXXCmd(args [][]bytes) 				// execute command
*/

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"
)

type Server struct {
	Conf  map[string]interface{}
	Store *Store
	CONNS int
}

func Logf(s string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("[%v] >> %s", time.Now().Format(time.StampMilli), s), args...)
}

func GetServer(conf string) *Server {
	log.Printf("LOADING %q\n", conf)
	b, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Panic(err)
	}
	log.Println("CONFIGURING SERVER")
	var confM map[string]interface{}
	if err := json.Unmarshal(b, &confM); err != nil {
		log.Panic(err)
	}
	if greet, ok := confM["banner"].(string); ok {
		log.Println(greet)
	}
	return &Server{
		Conf:  confM,
		Store: GetStore(int64(confM["gcrate"].(float64))),
	}
}

func (self *Server) ListenAndServe() {
	addr, err := net.ResolveTCPAddr("tcp", self.Conf["addr"].(string))
	if err != nil {
		log.Panicln(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Panicln(err)
	}
	log.Printf("LISTENING ON %q, ACCEPTING CONNECTIONS...\n", self.Conf["addr"])
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Panicln(err)
		}
		self.CONNS++
		log.Printf("ACCEPTED CLIENT REQUEST %q\n", conn.RemoteAddr())
		log.Printf("(TOTAL: %d)\n", self.CONNS)
		go self.HandleConn(conn)
	}
}

// handle connection
func (self *Server) HandleConn(conn *net.TCPConn) {
	r := bufio.NewReader(conn)
	TTL := self.Conf["ttl"].(float64)
	self.extendTTL(conn, TTL)
	for {
		b, err := r.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			self.closeConn(conn)
			return
		} else {
			self.extendTTL(conn, TTL)
		}
		args := bytes.Split(bytes.ToLower(bytes.TrimRight(b, "\r\n")), []byte(" "))
		switch string(args[0]) {
		case "ping":
			conn.Write([]byte("PONG\r\n"))
			break
		case "set":
			if len(args[1:]) == 2 {
				ok := self.Store.Set(string(args[1]), args[2])
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "app":
			if len(args[1:]) >= 2 {
				ok := self.Store.App(string(args[1]), args[2:]...)
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "get":
			if len(args[1:]) == 1 {
				v := self.Store.Get(string(args[1]))
				v = append(v, []byte("\r\n"))
				conn.Write(bytes.Join(v, []byte(" ")))
			}
			break
		case "del":
			if len(args[1:]) == 1 {
				ok := self.Store.Del(string(args[1]))
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "exp":
			if len(args[1:]) == 2 {
				n, _ := strconv.Atoi(string(args[2]))
				ok := self.Store.Exp(string(args[1]), int64(n))
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "ttl":
			if len(args[1:]) == 1 {
				v := self.Store.TTL(string(args[1]))
				n := strconv.Itoa(v)
				conn.Write([]byte(n + "\r\n"))
			}
			break
		case "haskey":
			if len(args[1:]) == 1 {
				ok := self.Store.HasKey(string(args[1]))
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "getval":
			if len(args[1:]) >= 2 {
				var ns []int
				for i := 2; i < len(args[2:]); i++ {
					n, _ := strconv.Atoi(string(args[i]))
					ns = append(ns, n)
				}
				v := self.Store.GetVal(string(args[1]), ns...)
				v = append(v, []byte("\r\n"))
				conn.Write(bytes.Join(v, []byte(" ")))
			}
			break
		case "delval":
			if len(args[1:]) == 2 {
				ok := self.Store.DelVal(string(args[1]), args[2])
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "save":
			if len(args[1:]) == 1 {
				ok := self.Store.SaveSnapshot(string(args[1]))
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "load":
			if len(args[1:]) == 1 {
				ok := self.Store.LoadSnapshot(string(args[1]))
				conn.Write([]byte(boolToString(ok) + "\r\n"))
			}
			break
		case "purge":
			ok := self.Store.Purge()
			conn.Write([]byte(boolToString(ok) + "\r\n"))
			break
		case "exit":
			conn.SetDeadline(time.Now())
			break
		}
	}
}

// close connection
func (self *Server) closeConn(conn *net.TCPConn) {
	conn.Write([]byte("CLOSING CONNECTION, GOODBYE\r\n"))
	log.Printf("CLOSED CONNECTION TO CLIENT [%s]\n", conn.RemoteAddr().String())
	conn.Close()
	conn = nil
}

// extend conn ttl
func (self *Server) extendTTL(conn *net.TCPConn, ttl float64) {
	if ttl > 0 {
		conn.SetDeadline(time.Now().Add(time.Duration(ttl) * time.Second))
	}
}
