//go:build integration

// Copyright (c) 2023 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package integration_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tcmsoftware/mirroring-web-crawler/crawler"
)

func TestCrawler(t *testing.T) {
	const startUrl = "https://blog.cleancoder.com/"
	const destDir = "saved"
	const timeout = 10 * time.Second
	log := log.New(os.Stdout, "INTEGRATION TEST :", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	c := crawler.New(startUrl, destDir, timeout, log)
	if err := c.Run(); err != nil {
		t.Fatalf("error when running crawler: %v", err)
	}
	modTimes := make(map[string]time.Time, 200)
	var count int
	err := filepath.Walk(destDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			count++
			modTimes[path] = info.ModTime()
			return nil
		})
	require.Nil(t, err)
	require.True(t, count > 0)

	// Crawling same page again; files will not be overwritten
	// and not be downloaded again, so their modification times
	// will be the same.
	time.Sleep(1 * time.Second)
	modTimesForSecondRun := make(map[string]time.Time, 200)
	if err := c.Run(); err != nil {
		t.Fatalf("error when running crawler: %v", err)
	}
	err = filepath.Walk(destDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			count++
			modTimesForSecondRun[path] = info.ModTime()
			return nil
		})
	require.Nil(t, err)
	require.Equal(t, modTimes, modTimesForSecondRun)
}
