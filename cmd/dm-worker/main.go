// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pingcap/dm/dm/worker"
	"github.com/pingcap/dm/pkg/log"
	"github.com/pingcap/dm/pkg/utils"
	"github.com/pingcap/errors"
)

func main() {
	cfg := worker.NewConfig()
	err := cfg.Parse(os.Args[1:])
	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Errorf("parse cmd flags err %s", err)
		os.Exit(2)
	}

	log.SetLevelByString(strings.ToLower(cfg.LogLevel))
	if len(cfg.LogFile) > 0 {
		log.SetOutputByName(cfg.LogFile)
	}

	utils.PrintInfo("worker", func() {
		log.Infof("config: %s", cfg)
	})

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	s := worker.NewServer(cfg)

	go func() {
		sig := <-sc
		log.Infof("got signal [%v] to exit", sig)
		s.Close()
	}()

	err = s.Start()
	if err != nil {
		log.Errorf("start dm-worker err %s", err)
		os.Exit(2)
	}
	s.Close() // wait until closed
	log.Info("dm-worker exit")
}
