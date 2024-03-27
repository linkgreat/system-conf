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
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"system-conf/common/log"
	"time"
)

func ParseIntFromQuery(c *gin.Context, key string) (result *int) {
	if v, ok := c.GetQuery(key); ok {
		if vv, err := strconv.Atoi(v); err == nil {
			result = &vv
		}
	}
	return
}
func ParseUintFromQuery(c *gin.Context, key string) (result *uint) {
	if v, ok := c.GetQuery(key); ok {
		if vv, err := strconv.ParseUint(v, 10, 64); err == nil {
			x := uint(vv)
			result = &x
		}
	}
	return
}
func ParseIntFromDefaultQuery(c *gin.Context, key, defaultVal string) (result *int) {
	if v := c.DefaultQuery(key, defaultVal); len(v) > 0 {
		if vv, err := strconv.Atoi(v); err == nil {
			result = &vv
		}
	}
	return
}
func ParseIntFromQueryV0(c *gin.Context, key string) (result int, err error) {
	if v, ok := c.GetQuery(key); ok {
		if result, err = strconv.Atoi(v); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("key:%s not found", key)
	}
	return
}
func ParseUintFromPath(c *gin.Context, key string) (result *uint) {
	if v := c.Param(key); len(v) > 0 {
		if vv, err := strconv.Atoi(v); err == nil {
			x := uint(vv)
			result = &x
			return
		}
	}
	return
}
func ParseIntFromPath(c *gin.Context, key string) (result *int) {
	if v := c.Param(key); len(v) > 0 {
		if vv, err := strconv.Atoi(v); err == nil {
			result = &vv
			return
		}
	}
	return
}

func ParseTmFromString(tmStr string) (tm *time.Time) {
	vv, err := time.ParseInLocation(time.RFC3339, tmStr, log.BJ)
	if err != nil {
		if vv, err = time.ParseInLocation("2006-01-02T15:04:05", tmStr, log.BJ); err != nil {
			if vv, err = time.ParseInLocation("2006-01-02T15-04-05", tmStr, log.BJ); err != nil {
				if vv, err = time.ParseInLocation("2006-01-02 15:04:05", tmStr, log.BJ); err != nil {
					if vv, err = time.ParseInLocation("2006-01-02-15-04-05", tmStr, log.BJ); err != nil {
						if vv, err = time.ParseInLocation("2006-01-02", tmStr, log.BJ); err != nil {
							return
						}
					}
				}
			}
		}
	}
	if err == nil {
		tm = &vv
	}
	return
}
func ParseTmFromQuery(c *gin.Context, key string) (tm *time.Time) {
	if v, ok := c.GetQuery(key); ok {
		tm = ParseTmFromString(v)
	}
	return
}

func ParseTmFromQuery0(c *gin.Context, key string) (tm time.Time, err error) {
	if v, ok := c.GetQuery(key); ok {
		if tm, err = time.ParseInLocation(time.RFC3339, v, log.BJ); err != nil {
			if tm, err = time.ParseInLocation("2006-01-02T15:04:05", v, log.BJ); err != nil {
				if tm, err = time.ParseInLocation("2006-01-02 15:04:05", v, log.BJ); err != nil {
					if tm, err = time.ParseInLocation("2006-01-02-15-04-05", v, log.BJ); err != nil {
						if tm, err = time.ParseInLocation("2006-01-02", v, log.BJ); err != nil {
							return
						}
					}
				}
			}
		}
	} else {
		err = fmt.Errorf("query key:%s not found", key)
	}
	return
}
func ParseTmDurationFromQuery(c *gin.Context, from, to string) (duration time.Duration) {
	var tm0, tm1 time.Time
	if v := ParseTmFromQuery(c, from); v != nil {
		tm0 = *v
	}
	if v := ParseTmFromQuery(c, to); v != nil {
		tm1 = *v
	}
	if !tm0.IsZero() {
		if tm1.IsZero() {
			duration = time.Now().Sub(tm0)
		} else {
			duration = tm1.Sub(tm0)
		}
	} else {
		if tm1.IsZero() {
			tm, _ := time.ParseInLocation("2006-01-02", "2020-01-01", log.BJ)
			duration = tm1.Sub(tm)
		}
	}
	return
}

func ParseQuery2IntArray(c *gin.Context, qName, sep string) []int {
	if x, ok := c.GetQuery(qName); ok {
		array := make([]int, 0)
		for _, s := range strings.Split(x, sep) {
			if xx, err := strconv.Atoi(s); err == nil {
				array = append(array, xx)
			}
		}
		return array
	}
	return nil
}
func ParseQuery2UintArray(c *gin.Context, qName, sep string) (array []uint) {
	if x, ok := c.GetQuery(qName); ok {
		for _, s := range strings.Split(x, sep) {
			if xx, err := strconv.Atoi(s); err == nil {
				array = append(array, uint(xx))
			}
		}
	}
	return
}
func ParseQuery2FloatArray(c *gin.Context, qName, sep string) []float64 {
	if x, ok := c.GetQuery(qName); ok {
		array := make([]float64, 0)
		for _, s := range strings.Split(x, sep) {
			if x, err := strconv.ParseFloat(s, 64); err == nil {
				array = append(array, x)
			}
		}
		return array
	}
	return nil
}

func ParseQuery2Strs(c *gin.Context, qName, sep string) []string {
	if x, ok := c.GetQuery(qName); ok {
		array := make([]string, 0)
		for _, s := range strings.Split(x, sep) {
			if len(s) > 0 {
				array = append(array, s)
			}
		}
		return array
	}
	return nil
}
func Abort(c *gin.Context, sts int, err error, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	code := sts
	if sts >= http.StatusOK && sts <= http.StatusAlreadyReported {
		code = 0
	} else {
		code = sts
	}

	if err != nil {
		err = fmt.Errorf("%v; err:%v", msg, err)
		msg = err.Error()
	}
	data := gin.H{
		"code": code,
		"msg":  msg,
	}
	c.Set("r", data)
	c.AbortWithStatusJSON(sts, data)
}
func GetJwtToken(c *gin.Context) (token string) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
	}

	// 检查 authHeader 是否为 Bearer 类型
	if !strings.HasPrefix(authHeader, "Bearer ") {
	}

	// 提取 token 部分
	token = strings.TrimPrefix(authHeader, "Bearer ")
	return token
}
func MiddlewareParseToken(key []byte, tokenFieldName string, auth bool, ignoredPaths ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		for _, p := range ignoredPaths {
			if c.Request.URL != nil && p == c.Request.URL.Path {
				return
			}
		}
		tokenKey := tokenFieldName
		if v, ok := c.Keys[tokenKey].(string); ok {
			tokenKey = v
		}
		token, err := c.Cookie(tokenFieldName)
		if err != nil {

			if token = GetJwtToken(c); len(token) == 0 {
				if len(token) == 0 || token == "undefined" {
					if t, ok := c.GetQuery("token"); ok {
						token = t
					}
				}
				if len(token) == 0 || token == "undefined" {
					token = c.GetHeader(tokenFieldName)
					if len(token) == 0 || token == "undefined" {
						token = c.GetHeader("Authorization")
						if len(token) == 0 || token == "undefined" {
							token = c.GetHeader("x-auth-token")
							if len(token) == 0 || token == "undefined" {
								if t, ok := c.GetQuery(tokenKey); ok {
									token = t
								}

							}
						}
					}
				}
			}

		}
		if len(token) > 0 {
			c.Set("jwt", token)
		}
		tk, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return key, nil
		})
		c.Set("token_raw", tk)
		if err != nil || tk == nil { //|| !tk.Valid {
			if auth {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}
		if tkc, ok := tk.Claims.(jwt.MapClaims); ok {
			if err := tkc.Valid(); err != nil {

				Abort(c, http.StatusUnauthorized, err, "bad token")
				return
			}
			c.Set(tokenKey, tkc)
			if v, ok := tkc["domains"].([]interface{}); ok {
				c.Set("domains", v)
			}
			if v, ok := tkc["id"]; ok {
				c.Set("uid", v)
			}
			//c.Set("uid", tkc["uid"])
		}
	}
}
