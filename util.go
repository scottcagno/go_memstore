// -------
// util.go ::: utilities
// -------
// Copyright (c) 2013-Present, Scott Cagno. All rights reserved.
// This source code is governed by a BSD-style license.

package data

import ()

// marshal bool to string
func boolToString(ok bool) string {
	if ok {
		return "OK"
	}
	return "ERR"
}

// arg parser
type ArgParser struct {
	args     [][]byte
	argc     int
	rule_cnt int
	rule_exp []int
}

const (
	CMD = 0
	KEY = 1
	VAL = 2
	WLD = 3
)

// return arg parser instance
func ArgParse(args [][]byte) *ArgParser {
	return &ArgParser{
		args: args,
		argc: len(args),
	}
}

// set x testing123
func (self *ArgParser) Rules(v ...int) *ArgParser {
	self.rule_exp = append(self.rule_exp, v...)
	self.rule_cnt = len(self.rule_exp)
	return self
}

func (self *ArgParser) Parse() [][]byte {
	if self.argc >= self.rule_cnt {
		for i := range self.rule_exp {
			switch self.rule_exp {
			case CMD:
				_, ok := self.args[i]
			}
		}
	}
	return nil
}
