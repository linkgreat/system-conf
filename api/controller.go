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
	"github.com/gin-gonic/gin"
	"reflect"
	"strings"
)

func BindHandler(this any, funcPrefix string, parent gin.IRouter) {
	val := reflect.TypeOf(this)
	sz := val.NumMethod()
	for i := 0; i < sz; i++ {
		method := val.Method(i)
		// fmt.Println("got method", method.Name)
		if strings.HasPrefix(method.Name, funcPrefix) {
			method.Func.Call([]reflect.Value{reflect.ValueOf(this), reflect.ValueOf(parent)})
		}
	}
}

type Controller struct {
	Parent gin.IRouter
}

func NewController(parent gin.IRouter) *Controller {
	ctx := &Controller{
		Parent: parent,
	}

	return ctx
}
