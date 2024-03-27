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

package log

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/unknwon/goconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// error logger
var myLogger *zap.SugaredLogger

//go:embed Shanghai
var shanghai []byte
var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

const (
	tm_fmt = "2006-01-02 15:04:05.000"
)

var Loc *time.Location
var ProcName string
var Sn string
var BJ *time.Location
var Silent = false

func IsSilent() bool {
	if v := os.Getenv("LOG_SILENT"); v == "1" {
		return true
	}
	return Silent
}
func init() {
	var loc *time.Location
	if v := os.Getenv("LOG_SILENT"); v == "1" {
		Silent = true
	}
	if v := os.Getenv("LOG_SILENT"); v == "1" {

	}
	loc, err := time.LoadLocationFromTZData("Shanghai", shanghai)
	if err != nil {
		fmt.Printf("failed to load tz data form static resource. err:%v\n", err)
	} else {
		Loc = loc
		BJ = loc
	}
	if v := os.Getenv("LOG_DEBUG"); v == "1" {
		fmt.Printf("got location: %s\n", loc.String())
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "linenum",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder, //控制台彩色日志输出
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(loc).Format(tm_fmt))
		},
		//EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), //时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, // 时间精度？
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	//windows下大写DEBUG等标签展示
	if runtime.GOOS == "windows" {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	var pathInfo = struct {
		DefaultDir string `json:"defaultDir"`
		WorkDir    string `json:"workDir"`
		BaseDir    string `json:"baseDir"`
		Actual     string `json:"actual"`
	}{
		DefaultDir: path.Dir(os.Args[0]),
	}
	softDir := pathInfo.DefaultDir
	if dir, e := os.Getwd(); e == nil {
		pathInfo.WorkDir = dir
		softDir = dir
	}
	if x := os.Getenv("BaseDir"); x != "" {
		pathInfo.BaseDir = x
		softDir = x
	}
	pathInfo.Actual = softDir
	tmpBin, _ := json.Marshal(pathInfo)
	if !IsSilent() {
		fmt.Printf("current log path info: %s\n", string(tmpBin))
	}

	// 设置日志级别
	lvl := "info"
	if cfg, err := goconfig.LoadConfigFile(softDir + "/conf.ini"); err == nil {
		lvl, _ = cfg.GetValue("config", "log_level")
	}
	if v := os.Getenv("LOG_LEVEL"); len(v) > 0 {
		lvl = v
	}

	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(getLoggerLevel(lvl))
	var writeSyncer zapcore.WriteSyncer
	if IsSilent() {
		writeSyncer = zapcore.NewMultiWriteSyncer()
	} else if v := os.Getenv("LOG_NOFILE"); len(v) == 0 || v == "0" {
		ProcName = filepath.Base(os.Args[0])
		extName := filepath.Ext(ProcName)
		if extName != "" {
			ProcName = strings.ReplaceAll(ProcName, extName, "")
		}
		//if len(os.Args)>1{
		//	procName = fmt.Sprintf("%s_%s", procName, os.Args[1])
		//}
		logSubPath := ProcName

		for _, a := range os.Args[1:] {
			if !strings.HasPrefix(a, "-") {
				logSubPath = path.Join(logSubPath, a)
				ProcName += fmt.Sprintf("_%s", a)
			}
		}
		pid := os.Getpid()
		logPath := path.Join(softDir, "logs", logSubPath)
		fileName := fmt.Sprintf("%s/%s_%s_%d.log", logPath, ProcName, time.Now().In(loc).Format("20060102_150405"), pid)
		fileName = fmt.Sprintf("%s/%s.log", logPath, ProcName)
		if vv := os.Getenv("LOG_DEBUG"); vv == "1" {
			fmt.Println("当前日志文件：", fileName)
		}
		hook := lumberjack.Logger{
			Filename:   fileName, // 日志文件路径
			MaxSize:    1,        // 每个日志文件保存的最大尺寸 单位：M
			MaxBackups: 500,      // 日志文件最多保存多少个备份
			MaxAge:     60,       // 文件最多保存多少天
			Compress:   false,    // 是否压缩
		}
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // 日志格式
		writeSyncer,                              // 打印到控制台和文件
		atomicLevel,                              // 日志级别
	)

	//日志级别=debug时，
	if lvl == "debug" {
		caller := zap.AddCaller()           //开启开发模式，堆栈跟踪
		development := zap.AddCallerSkip(1) //开启文件及行号
		logger := zap.New(core, caller, development)
		myLogger = logger.Sugar()
	} else {
		logger := zap.New(core)
		myLogger = logger.Sugar()
	}
}

// 兼容 log.Println [INFO]级别
func Println(args ...interface{}) {
	myLogger.Info(args...)
}

// 兼容 log.Printf [INFO]级别
func Printf(template string, args ...interface{}) {
	myLogger.Infof(template, args...)
}

func Debug(args ...interface{}) {
	myLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	myLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	myLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	myLogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	myLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	myLogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	myLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	myLogger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	myLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	myLogger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	myLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	myLogger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	myLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	myLogger.Fatalf(template, args...)
}
