package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/fatih/set.v0"
)

var (
	taskIter       *Iter
	taskQueue      chan map[string]string
	workerPool     WorkerPool
	threadNum      int
	timeout        int
	vulFile        string
	verifyFunc     func(map[string]string)
	deadHost       *set.Set
	vulHost        *set.Set
	LF             = GetLF()
	writeLock      sync.Mutex
	outputStr      string
	outputLen      int
	cycle          bool
	hostConnNumMap = HostConnNumMap{m: map[string]int{}}
	hostConnNum    int
)

func init() {
	var (
		hostFile string
		userFile string
		passFile string
		mod      string
	)

	hostFile, userFile, passFile, vulFile, mod, threadNum, timeout, cycle = parseArgs()

	hosts, err := ReadLines(hostFile)
	CheckErr(err)
	usernames, err := ReadLines(userFile)
	CheckErr(err)
	passwords, err := ReadLines(passFile)
	CheckErr(err)

	deadHost = set.New()
	vulHost = set.New()
	keyList := []string{"username", "password", "host"}
	listMap := map[string][]string{
		"username": usernames,
		"password": passwords,
		"host":     hosts,
	}
	taskIter = &Iter{}
	taskIter.Init(listMap, keyList)

	if threadNum == 0 {
		threadNum = 100
	}

	switch mod {
	case "ssh":
		verifyFunc = sshLogin
		hostConnNum = 5
	case "mysql":
		verifyFunc = mysqlLogin
		hostConnNum = 4
	default:
		FlushPrint(red("Unknown model!" + LF))
		os.Exit(1)
	}

	taskQueue = make(chan map[string]string, 100)
}

func parseArgs() (string, string, string, string, string, int, int, bool) {
	hostFile := flag.String("h", "host.txt", "Host file, one host per line, such as 1.1.1.1:22")
	userFile := flag.String("u", "username.txt", "Username file, one username per line")
	passFile := flag.String("p", "password.txt", "Password file, one password per line")
	vulFile := flag.String("o", "vul.txt", "Output file")
	mod := flag.String("m", "ssh", "Model, select from ssh, mysql")
	threadNum := flag.Int("t", 100, "Thread number")
	timeout := flag.Int("T", 5, "Timeout for host")
	cycle := flag.Bool("c", false, "Monitor model")
	flag.Parse()
	return *hostFile, *userFile, *passFile, *vulFile, *mod, *threadNum, *timeout, *cycle
}

func worker(id int) {
	for {
		select {
		case taskItem := <-taskQueue:
			workerPool.FreePop()
			if !deadHost.Has(taskItem["host"]) && !vulHost.Has(taskItem["host"]) {
				if hostConnNumMap.GetCount(taskItem["host"]) < hostConnNum {
					hostConnNumMap.AddCount(taskItem["host"])
					verifyFunc(taskItem)
					hostConnNumMap.DoneCount(taskItem["host"])
				} else {
					taskQueue <- taskItem
				}
			} else {
				hostConnNumMap.DelHost(taskItem["host"])
			}
			workerPool.FreeAdd()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func addTask() {
	for {
		if taskItem := taskIter.Next(); taskItem != nil {
			taskQueue <- taskItem
			continue
		}
		break
	}
}

func main() {
	workerPool = WorkerPool{}
	workerPool.Init(worker, threadNum)
	if cycle {
		go addTask()
		for {
			text := MakeBar(20, taskIter.Percent()) + outputStr
			FlushPrint(PrepareText(text))
			if workerPool.freeNum > workerPool.workerNum-10 && workerPool.freeNum < workerPool.workerNum {
				FlushPrint(red(PrepareText(fmt.Sprintf("x: %d total: %d workerPool.freeNum: %d", taskIter.x, taskIter.Total, workerPool.freeNum)) + LF))
			}
			time.Sleep(1 * time.Second)
			if workerPool.freeNum == workerPool.workerNum && len(taskQueue) == 0 {
				deadHost.Clear()
				vulHost.Clear()
				taskIter.Reset()
				FlushPrint(magenta(PrepareText(fmt.Sprintf("[%s]finished", NowTime(""))) + LF))
				Write2File(vulFile, "["+NowTime("")+"]finished"+LF)
				os.Remove(vulFile + ".report")
				os.Rename(vulFile, vulFile+".report")
				FlushPrint(magenta(PrepareText(fmt.Sprintf("[%s]Sleeping...", NowTime(""))) + LF))
				time.Sleep(15 * time.Second)
				go addTask()
			}
		}
	} else {
		go addTask()
		for {
			text := MakeBar(20, taskIter.Percent()) + outputStr
			FlushPrint(PrepareText(text))
			time.Sleep(1 * time.Second)
			if taskIter.Percent() == 100.0 {
				FlushPrint(yellow(PrepareText(fmt.Sprintf("x: %d total: %d workerPool.freeNum: %d", taskIter.x, taskIter.Total, workerPool.freeNum)) + LF))
			}
			if workerPool.freeNum == workerPool.workerNum && len(taskQueue) == 0 {
				FlushPrint(RightPad(MakeBar(20, taskIter.Percent()), outputLen, " "))
				workerPool.Close()
				break
			}
		}
	}
	workerPool.Wait()
}
