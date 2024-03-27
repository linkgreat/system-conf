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
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type IPagination interface {
	GetTotal() int
	GetData() any
}
type PageSummary struct {
	Total int `bson:"total" json:"total"`
}
type PageResult[T any] struct {
	Summary []PageSummary `bson:"summary" json:"summary"`
	Data    []T           `bson:"data" json:"data"`
}

func (m *PageResult[T]) GetTotal() int {
	if len(m.Summary) > 0 {
		return m.Summary[0].Total
	}
	return 0
}

func (m *PageResult[T]) GetData() any {
	return m.Data
}

type Response struct {
	Code      int         `json:"code" example:"0"`
	Message   string      `json:"msg" example:"ok"`
	CreatedId any         `json:"createdId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Total     interface{} `json:"total,omitempty"`
	Summary   interface{} `json:"summary,omitempty"`
	DebugInfo interface{} `json:"debugInfo,omitempty"`
}

func NewRestResponse() *Response {
	return &Response{}
}
func (resp *Response) SetMessage(msg string, args ...interface{}) *Response {
	resp.Message = fmt.Sprintf(msg, args...)
	return resp
}
func (resp *Response) SetCode(code int) *Response {
	resp.Code = code
	return resp
}
func (resp *Response) SetData(data interface{}) *Response {
	resp.Data = data
	return resp
}
func (resp *Response) SetSummary(data interface{}) *Response {
	resp.Summary = data
	return resp
}
func (resp *Response) SetDebugInfo(data interface{}) *Response {
	resp.DebugInfo = data
	return resp
}
func (resp *Response) SetTotal(data interface{}) *Response {
	resp.Total = data
	return resp
}

func (resp *Response) Abort(c *gin.Context, status int) *Response {
	if resp.Code == 0 {
		resp.Code = status
	}
	c.AbortWithStatusJSON(status, resp)
	if v, ok := c.Keys["debug"].(int); ok {
		if v > 0 {
			fmt.Printf("resp message:%s\n", resp.Message)
		}
	}
	return resp
}
func (resp *Response) Created(c *gin.Context, id any) *Response {
	resp.CreatedId = id
	c.JSON(http.StatusCreated, resp)
	return resp
}
func (resp *Response) OK0(c *gin.Context) *Response {
	if len(resp.Message) == 0 {
		resp.Message = "ok"
	}
	c.JSON(http.StatusOK, resp)
	return resp
}
func (resp *Response) OK(c *gin.Context) *Response {
	if len(resp.Message) == 0 {
		resp.Message = "ok"
	}
	buf, _ := json.Marshal(resp)
	//c.Stream(func(w io.Writer) bool {
	//	enc := json.NewEncoder(w)
	//	if err := enc.Encode(resp); err != nil {
	//		return false
	//	}
	//	return true
	//})

	c.Data(http.StatusOK, "application/json", buf)
	return resp
}
