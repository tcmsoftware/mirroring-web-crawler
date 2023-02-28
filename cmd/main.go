// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/tcmsoftware/mirroring-web-crawler/crawler"
)

func run(log *log.Logger, startUrl, destDir string) error {
	log.Println("main: starting web crawler")
	defer log.Println("main: completed")
	const defaultTimeOut = 10 * time.Second
	if startUrl == "" {
		return errors.New("missing start url")
	}
	if destDir == "" {
		return errors.New("missing dest dir")
	}
	c := crawler.New(startUrl, destDir, defaultTimeOut, log)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	crawlerErrors := make(chan error, 1)
	go func() {
		crawlerErrors <- c.Run()
	}()
	select {
	case err := <-crawlerErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		log.Printf("main: %v: Start shutdown", sig)
	}
	return nil
}

func main() {
	log := log.New(os.Stdout, "WEB CRAWLER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	startUrl := flag.String("u", "", "url")
	destDir := flag.String("d", "", "dest dir")
	flag.Parse()
	if err := run(log, *startUrl, *destDir); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
