package xfmt

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Printf 为控制台打印附着时间和caller，并且着色
func Printf(format string, a ...any) {
	var buf bytes.Buffer
	prefix := fmt.Sprintf("%-32s %-17s", color.YellowString(time.Now().Format("2006-01-02 15:04:05.999")), color.GreenString("INFO"))
	buf.WriteString(prefix)
	buf.WriteString(fmt.Sprintf("%-40s", addShortCaller()))
	buf.WriteString(format)
	buf.WriteString("\n")
	_, _ = fmt.Fprintf(os.Stdout, buf.String(), a...)
}

func addShortCaller() string {
	_, file, line, _ := runtime.Caller(2)
	paths := strings.Split(file, "/")
	short := strings.Join(paths[len(paths)-2:], "/")
	return fmt.Sprintf("%s:%d", short, line)
}
