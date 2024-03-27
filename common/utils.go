/*
 * Copyright (c) 2023 fjw
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package common

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"system-conf/common/log"
)

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

var SkipVerify = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

func WaitParentExit() {
	WaitParent(func() {
		fmt.Printf("will be terminated because parent exited. \n")
		syscall.Exit(-1)
		os.Exit(-1)
	})
}
func WaitParent(cb func()) {
	ppid := os.Getppid()
	if proc, err := os.FindProcess(ppid); err == nil {
		log.Printf("parent pid = %v\n", ppid)
		_, err = proc.Wait()
		if err != nil {
			log.Printf("failed to wait parent. err:%v\n", err)
		} else {
			if cb != nil {
				cb()
			}
		}

	}
}
func GetParentUse(c *cobra.Command) string {
	if c.Parent() != nil {
		return GetParentUse(c.Parent()) + "." + c.Use
	} else {
		return c.Use
	}
}
func ParseEventSource(c *cobra.Command) string {
	result := filepath.Base(os.Args[0])
	if c != nil {
		result = GetParentUse(c)
	}
	fmt.Println("current event source:", result)
	return result
}
