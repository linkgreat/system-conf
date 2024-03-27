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

package api

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"system-conf/common"
	"system-conf/common/log"
	"time"

	"net/http"
	"os"
	"os/exec"
)

//go:embed ip.conf
var ipConf string

func (m *Controller) AutoBindSystem() {
	sys := m.Parent.Group("/system")

	BindHandler(m, "BindSystemHandle", sys)
}

// BindSystemHandleGetSystemTm godoc
// @Summary 获取系统时间
// @Description 获取系统时间
// @Tags 系统
// @Security Bearer
// @Produce  json
// @Param tm query string true "时间" default(2023-03-27-15-04-05)
// @Success 200 {object} Response  '{"code":200,"data":[],"msg":"OK"}'
// @Router /system/time [get]
func (m *Controller) BindSystemHandleGetSystemTm(parent gin.IRouter) {
	parent.GET("/time", func(c *gin.Context) {
		resp := NewRestResponse()
		resp.SetData(time.Now().In(log.BJ).Format("2006-01-02 15:04:05")).OK(c)
	})
}

// BindSystemHandleUpdateTm godoc
// @Summary 更新系统时间
// @Description 更新系统时间
// @Tags 系统
// @Security Bearer
// @Produce  json
// @Param tm query string true "时间" default(2023-03-27-15-04-05)
// @Success 200 {object} Response  '{"code":200,"data":[],"msg":"OK"}'
// @Router /system/update.time [put]
func (m *Controller) BindSystemHandleUpdateTm(parent gin.IRouter) {
	parent.Any("/update.time", func(c *gin.Context) {
		resp := NewRestResponse()
		tm := common.ParseTmFromQuery(c, "tm")
		if tm == nil {
			resp.SetMessage("日期格式错误").Abort(c, http.StatusBadRequest)
			return
		}
		os.Setenv("TZ", log.BJ.String())
		fmt.Printf("current TZ=%s\n", os.Getenv("TZ"))

		cmd0 := exec.Command("date", "-s", tm.Format("2006-01-02 15:04:05"))
		output0, err := cmd0.Output()
		if err != nil {
			resp.SetMessage("修改时钟失败:%v", err).Abort(c, http.StatusInternalServerError)
			return
		} else {
			cmd1 := exec.Command("hwclock", "--systohc")
			if output1, err1 := cmd1.Output(); err1 != nil {
				resp.SetMessage("同步时钟失败:%v", err1).Abort(c, http.StatusInternalServerError)
				return
			} else {
				resp.SetMessage(string(output0) + "\n" + string(output1)).OK(c)
			}
		}
	})
}

// BindSystemHandleGetIp godoc
// @Summary 读取系统IP
// @Description 读取系统IP
// @Tags 系统
// @Security Bearer
// @Produce  json
// @Param ip query string true "ip" default(192.168.0.193)
// @Param mask query number true "掩码" default(24)
// @Success 200 {object} Response  '{"code":200,"data":[],"msg":"OK"}'
// @Router /system/ip [get]
func (m *Controller) BindSystemHandleGetIp(parent gin.IRouter) {
	parent.GET("/ip", func(c *gin.Context) {
		resp := NewRestResponse()
		args := []string{
			"-c",
			"ip", "addr", "|", "grep", "secondary", "|", "awk", "'{print $2}'",
		}
		cmd := exec.Command("sh", args...)
		output, err := cmd.Output()
		if err != nil {
			resp.SetMessage("获取ip失败:%v", err.Error()).Abort(c, http.StatusBadRequest)
			return
		} else {
			resp.SetData(string(output)).OK(c)
		}
	})
}

// BindSystemHandleChangeIp godoc
// @Summary 更新系统IP
// @Description 更新系统IP
// @Tags 系统
// @Security Bearer
// @Produce  json
// @Param ip query string true "ip" default(192.168.0.193)
// @Param mask query number true "掩码" default(24)
// @Success 200 {object} Response  '{"code":200,"data":[],"msg":"OK"}'
// @Router /system/change.ip [put]
func (m *Controller) BindSystemHandleChangeIp(parent gin.IRouter) {
	parent.Any("/change.ip", func(c *gin.Context) {
		resp := NewRestResponse()
		ip := ""
		if v, ok := c.GetQuery("ip"); ok {
			ip = v
		} else {
			resp.SetMessage("未指定ip地址").Abort(c, http.StatusBadRequest)
			return
		}
		mask := ""
		if v := common.ParseIntFromQuery(c, "mask"); v == nil {
			resp.SetMessage("mask格式不正确(0-32)").Abort(c, http.StatusBadRequest)
			return
		} else if *v < 0 || *v > 32 {
			resp.SetMessage("mask数字超限(0-32)").Abort(c, http.StatusBadRequest)
			return
		} else {
			mask = c.Query("mask")
		}
		conf := ipConf
		conf = strings.ReplaceAll(conf, "$$IP$$", ip)
		conf = strings.ReplaceAll(conf, "$$MASK$$", mask)
		confPath := "/etc/netplan/01-network-manager-all.yaml"
		var conf0 string
		if buf, e := os.ReadFile(confPath); e != nil {
			resp.SetMessage("读取原始文件失败:%v", e).Abort(c, http.StatusInternalServerError)
			return
		} else {
			conf0 = string(buf)
		}

		confBak := fmt.Sprintf(`/etc/netplan/01-network-manager-all.yaml.%s`, time.Now().In(log.BJ).Format("2006-01-02-15-04-05"))
		if e := os.WriteFile(confBak, []byte(conf0), os.ModePerm); e != nil {
			resp.SetMessage("备份文件写入失败:%v", e).Abort(c, http.StatusInternalServerError)
			return
		}

		var err error
		defer func() {
			if err != nil {
				if e := os.WriteFile(confPath, []byte(conf0), os.ModePerm); e != nil {
					log.Warnf("恢复配置文件失败：%v", e)
				}
			}
		}()
		if e := os.WriteFile(confPath, []byte(conf), os.ModePerm); e != nil {
			resp.SetMessage("写入配置失败:%v", e).Abort(c, http.StatusInternalServerError)
			return
		}
		cmd0 := exec.Command("netplan", "apply")
		var output []byte
		if output, err = cmd0.Output(); err != nil {
			resp.SetMessage("应用配置失败:%v", err).Abort(c, http.StatusBadRequest)
			return
		} else {
			resp.SetMessage(string(output)).OK(c)
		}
	})
}
