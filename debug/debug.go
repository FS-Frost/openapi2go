package debug

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func Sprintf(format string, a ...any) string {
	lineInfo := GetLineInfo()
	msg := fmt.Sprintf(format, a...)
	return fmt.Sprintf("[%s:%d] %s", lineInfo.FileName, lineInfo.LineNumber, msg)
}

func Errorf(format string, a ...any) error {
	lineInfo := GetLineInfo()
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("[%s:%d] %s", lineInfo.FileName, lineInfo.LineNumber, msg)
}

type FileInfo struct {
	FileName   string
	LineNumber int
}

func GetLineInfo() FileInfo {
	cwd, _ := os.Getwd()
	_, file, line, _ := runtime.Caller(2)
	file = strings.Replace(file, cwd, "", 1)
	if file[0] == '/' {
		file = file[1:]
	}

	return FileInfo{
		FileName:   file,
		LineNumber: line,
	}
}

func Dump(args ...any) {
	for _, a := range args {
		bs, _ := json.MarshalIndent(a, "", "  ")
		fmt.Println(string(bs))
	}
}
