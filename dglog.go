package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mgutz/ansi"
	"github.com/sirupsen/logrus"
)

type ColorScheme struct {
	InfoLevelStyle  string
	WarnLevelStyle  string
	ErrorLevelStyle string
	FatalLevelStyle string
	PanicLevelStyle string
	DebugLevelStyle string
	PrefixStyle     string
	TimestampStyle  string
}

type compiledColorScheme struct {
	InfoLevelColor  func(string) string
	WarnLevelColor  func(string) string
	ErrorLevelColor func(string) string
	FatalLevelColor func(string) string
	PanicLevelColor func(string) string
	DebugLevelColor func(string) string
	PrefixColor     func(string) string
	TimestampColor  func(string) string
}

var defaultColorScheme *ColorScheme = &ColorScheme{
	InfoLevelStyle:  "green",
	WarnLevelStyle:  "yellow",
	ErrorLevelStyle: "red",
	FatalLevelStyle: "red",
	PanicLevelStyle: "red",
	DebugLevelStyle: "blue",
	PrefixStyle:     "cyan",
	TimestampStyle:  "black+h",
}

func getCompiledColor(main string, fallback string) func(string) string {
	var style string
	if main != "" {
		style = main
	} else {
		style = fallback
	}
	return ansi.ColorFunc(style)
}

func compileColorScheme(s *ColorScheme) *compiledColorScheme {
	return &compiledColorScheme{
		InfoLevelColor:  getCompiledColor(s.InfoLevelStyle, defaultColorScheme.InfoLevelStyle),
		WarnLevelColor:  getCompiledColor(s.WarnLevelStyle, defaultColorScheme.WarnLevelStyle),
		ErrorLevelColor: getCompiledColor(s.ErrorLevelStyle, defaultColorScheme.ErrorLevelStyle),
		FatalLevelColor: getCompiledColor(s.FatalLevelStyle, defaultColorScheme.FatalLevelStyle),
		PanicLevelColor: getCompiledColor(s.PanicLevelStyle, defaultColorScheme.PanicLevelStyle),
		DebugLevelColor: getCompiledColor(s.DebugLevelStyle, defaultColorScheme.DebugLevelStyle),
		PrefixColor:     getCompiledColor(s.PrefixStyle, defaultColorScheme.PrefixStyle),
		TimestampColor:  getCompiledColor(s.TimestampStyle, defaultColorScheme.TimestampStyle),
	}
}

//日志自定义格式
type LogFormatter struct{}

//格式详情
func (s *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("20060102.15:04:05")
	var file string
	var line int
	if entry.Caller != nil {
		currExecFullDirPath := s.getCurrentAbPath()
		currFileFullPath := filepath.Dir(entry.Caller.File)
		currFile := filepath.Base(entry.Caller.File)
		currExecDirPath := currFileFullPath[len(currExecFullDirPath):]

		file = filepath.Join(currExecDirPath, currFile)[1:]
		line = entry.Caller.Line
	}

	level := s.getLevel(entry)
	msg := fmt.Sprintf("%s [%s] %-120s [%s:%d]\n", timestamp, level, entry.Message, file, line)
	return []byte(msg), nil
}

func (s *LogFormatter) getLevel(entry *logrus.Entry) string {
	defaultColorScheme := &ColorScheme{
		InfoLevelStyle:  "white",
		WarnLevelStyle:  "yellow",
		ErrorLevelStyle: "red",
		FatalLevelStyle: "red",
		PanicLevelStyle: "red",
		DebugLevelStyle: "blue",
		PrefixStyle:     "cyan",
		TimestampStyle:  "black+h",
	}

	colorScheme := compileColorScheme(defaultColorScheme)
	var levelColor func(string) string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = colorScheme.DebugLevelColor
	case logrus.InfoLevel:
		levelColor = colorScheme.InfoLevelColor
	case logrus.WarnLevel:
		levelColor = colorScheme.WarnLevelColor
	case logrus.ErrorLevel:
		levelColor = colorScheme.ErrorLevelColor
	case logrus.FatalLevel:
		levelColor = colorScheme.FatalLevelColor
	case logrus.PanicLevel:
		levelColor = colorScheme.PanicLevelColor
	default:
		levelColor = colorScheme.DebugLevelColor
	}

	level := levelColor(strings.ToUpper(entry.Level.String())[:1])

	return level
}

// 最终方案-全兼容
func (s *LogFormatter) getCurrentAbPath() string {
	dir := s.getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return s.getCurrentAbPathByCaller()
	}
	return dir
}

// 获取当前执行文件绝对路径
func (s *LogFormatter) getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func (s *LogFormatter) getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(new(LogFormatter))
	logrus.SetReportCaller(true)
}
