package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	printLock sync.Mutex
)

func PrepareText(text string) string {
	l := len(text)
	if l < outputLen {
		text = RightPad(text, outputLen, " ")
	}
	outputLen = l
	return text
}

func FlushPrint(text string) {
	printLock.Lock()
	defer printLock.Unlock()
	fmt.Fprint(color.Output, "\r"+text)
	os.Stdout.Sync()
}

func LeftPad(str string, length int, pad string) string {
	return strings.Repeat(pad, length-len(str)) + str
}

func RightPad(str string, length int, pad string) string {
	return str + strings.Repeat(pad, length-len(str))
}

func MakeBar(width int, percent float64) string {
	x := int(float64(width) * percent / 100)
	finished := x - 1
	unFinished := width - x
	currentStr := ">"
	if x == 0 {
		finished = 0
		currentStr = ""
	} else if x == width {
		currentStr = "="
	}
	return fmt.Sprintf("[%s%s] %.2f%%", strings.Repeat("=", finished)+currentStr, strings.Repeat("-", unFinished), percent)
}

func NowTime(timeType string) string {
	switch timeType {
	case "date":
		return time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	case "time":
		return time.Unix(time.Now().Unix(), 0).Format("03:04:05")
	default:
		return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 03:04:05")
	}
}

func GetLF() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	case "darwin":
		return "\r"
	default:
		return "\n"
	}
}

func Min(first int, rest ...int) int {
	min := first
	for _, v := range rest {
		if v < min {
			min = v
		}
	}
	return min
}

func CheckErr(err error) {
	if err != nil {
		fmt.Println(red(err.Error()))
		os.Exit(1)
	}
}

func ReadLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, strings.TrimRight(scanner.Text(), "\r\n"))
	}
	return lines, scanner.Err()
}

func Write2File(file string, text string) bool {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()
	f.WriteString(text)
	return true
}
