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
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var defaultFFPath = "./ffmpeg"

func init() {
	if runtime.GOOS == "windows" {
		defaultFFPath = "./ffmpeg.exe"
	}

}

type WebOptions struct {
	Port       int    `yaml:"port,omitempty" json:"port,omitempty" mapstructure:"port"`
	Tls        bool   `yaml:"tls,omitempty" json:"tls,omitempty" mapstructure:"tls"`
	CertName   string `yaml:"certName,omitempty" json:"certName,omitempty" mapstructure:"certName"`
	CertPath   string `yaml:"certPath,omitempty" json:"certPath,omitempty" mapstructure:"certPath"`
	Root       string `yaml:"root,omitempty" json:"root,omitempty" mapstructure:"root"`
	Link       string `yaml:"link,omitempty" json:"link,omitempty" mapstructure:"link"`
	Modules    string `yaml:"modules,omitempty" json:"modules,omitempty" mapstructure:"modules"`
	ShowDoc    bool   `yaml:"showDoc,omitempty" json:"showDoc,omitempty" mapstructure:"showDoc"`
	FfmpegPath string `yaml:"ffmpegPath"`
}

func (opts *WebOptions) Prepare(c *cobra.Command) {
	c.Flags().IntVar(&opts.Port, "web.port", 1010, "default http port")
	c.Flags().BoolVar(&opts.Tls, "web.tls", false, "if use tls")
	c.Flags().StringVar(&opts.CertName, "web.certname", "dxkjyjy.cn", "name of cert")
	c.Flags().StringVar(&opts.CertPath, "web.certpath", "/opt/certs", "path of cert")
	c.Flags().StringVar(&opts.Root, "web.root", "www", "default root dir")
	c.Flags().StringVar(&opts.Link, "web.link", "/var/www", "default link path")
	c.Flags().StringVar(&opts.Modules, "web.modules", "doc", "default module name list splited by ','")
	c.Flags().BoolVar(&opts.ShowDoc, "web.showDoc", false, "if show docs of api")

	c.Flags().StringVar(&opts.FfmpegPath, "ff-path", defaultFFPath, "default module name list splited by ','")
}

func (opts *WebOptions) Parse(bParse bool) {
	flag.IntVar(&opts.Port, "web.port", 1010, "default http port")
	flag.BoolVar(&opts.Tls, "web.tls", false, "if use tls")
	flag.StringVar(&opts.CertName, "web.certname", "dxkjyjy.cn", "name of cert")
	flag.StringVar(&opts.CertPath, "web.certpath", "/opt/certs", "path of cert")
	flag.StringVar(&opts.Root, "web.root", "www", "default mqtt root topic(pd=pile data)")
	flag.StringVar(&opts.Link, "web.link", "/var/www", "default link path")
	flag.StringVar(&opts.Modules, "web.modules", "doc", "default module name list splited by ','")
	flag.BoolVar(&opts.ShowDoc, "web.showDoc", false, "if show docs of api")
	flag.StringVar(&opts.FfmpegPath, "ff-path", defaultFFPath, "default module name list splited by ','")
	if bParse {
		flag.Parse()
	}
}
func (opts *WebOptions) SetBasePath(basepath string) {
	if !filepath.IsAbs(opts.Root) {
		opts.Root = filepath.Join(basepath, opts.Root)
		if !filepath.IsAbs(opts.Root) {
			if v, e := filepath.Abs(opts.Root); e != nil {
				opts.Root = v
			}
		}
	}
	if !filepath.IsAbs(opts.CertPath) {
		opts.CertPath = filepath.Join(basepath, opts.CertPath)
		if !filepath.IsAbs(opts.CertPath) {
			if v, e := filepath.Abs(opts.CertPath); e != nil {
				opts.CertPath = v
			}
		}
	}
	if !filepath.IsAbs(opts.Link) {
		opts.Link = filepath.Join(basepath, opts.Link)
		if !filepath.IsAbs(opts.Link) {
			if v, e := filepath.Abs(opts.Link); e != nil {
				opts.Link = v
			}
		}
	}
}
func (opts *WebOptions) SetStatic(router gin.IRouter) (err error) {
	tmp := strings.Split(opts.Modules, ",")
	errs := make([]error, 0)
	for _, mod := range tmp {
		if lp, e := opts.MakeLink(mod); e == nil {
			router.Static(mod, lp)
		} else {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		msg := ""
		for _, e := range errs {
			msg += fmt.Sprintf("%v\n", e)
		}
		err = errors.New(msg)
	}
	return
}

func (opts *WebOptions) MakeLink(moduleName string) (linkPath string, err error) {
	originPath := filepath.Join(opts.Root, moduleName)
	if runtime.GOOS != "windows" {
		linkPath = filepath.Join(opts.Link, moduleName)
		//if utils.Exists(linkPath) {
		if !Exists(opts.Link) {
			os.MkdirAll(opts.Link, os.ModePerm)
		}
		if err := os.RemoveAll(linkPath); err != nil {
			fmt.Printf("failed to delete path %v; err:%v\n", linkPath, err)
			//return
		}
		err = os.Symlink(originPath, linkPath)
		if err != nil {
			fmt.Printf("failed to use patch ui. err:%v\n", err)
		}
	} else {
		linkPath = originPath
	}
	return
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, East-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, East-CustomHeader, Keep-Alive, User-Agent, East-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		//if method == "OPTIONS" {
		//    c.JSON(http.StatusOK, "WebOptions Request!")
		//}
		if method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		// 处理请求
		c.Next() //  处理请求
	}
}
