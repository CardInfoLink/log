package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron"
)

const (
	QuickPay  = "quickpay"
	Phantom   = "phantom"
	AngryCard = "angrycard"
	LogDir    = "logs/"
	FileType  = ".log"
)

var LogFile *os.File = nil
var FileName string = ""

func init() {
	// 日志初始化
	PreLog()
	RemoveLogFile()
	c := cron.New()
	c.AddFunc("0 0 0 * * *", PreLog)        // 每隔一天
	c.AddFunc("0 0 0 1 * *", RemoveLogFile) // 每隔一个月执行
	c.Start()
}

// Printer 定义了打印接口
type Printer interface {

	// 所有方法最终归为这个方法，真正打印日志
	Tprintf(v, l Level, tag string, format string, m ...interface{})

	// ChangeFormat 改变日志格式
	ChangeFormat(format string)

	// ChangeWriter 改变输出流
	ChangeWriter(w io.Writer)
}

// 由于改方法是定时任务，第一次和程序一起启动，如果失败，在控制台输出错误信息，如果后续定时任务出现处理失败，则沿用原来的log文件
func PreLog() {
	if FileName == "" {
		sPath, err := os.Getwd()
		if err != nil {
			slog := fmt.Sprintf("get the current file path error, the error is %s\n", err)
			if LogFile == nil {
				fmt.Println(slog)
				os.Exit(1)
			}
			Error(slog)
			return
		}
		strArrays := strings.Split(sPath, "/")
		ilen := len(strArrays)
		if ilen == 0 {
			slog := fmt.Sprintf("parse the file path to array error, the filepath is %s\n", sPath)
			if LogFile == nil {
				fmt.Println(slog)
				os.Exit(1)
			}
			Error(slog)
			return
		}
		FileName = strArrays[ilen-1]
	}

	if (FileName != QuickPay) && (FileName != Phantom) && (FileName != AngryCard) {
		SetPrinter(NewStandard(os.Stdout, DefaultFormat))
		return
	}

	// 类似phantom_20160512.log与quickpay_20160512.log
	sName := LogDir + FileName + "_" + time.Now().Format("20060102") + FileType
	pFile, err := os.Create(sName)
	if (err != nil) || (pFile == nil) {
		slog := fmt.Sprintf("create the log file error, the error is %s\n", err)
		if LogFile == nil {
			fmt.Println(slog)
			os.Exit(1)
		}
		Error(slog)
		return
	}

	SetPrinter(NewStandard(pFile, DefaultFormat))
	if LogFile != nil {
		err = LogFile.Close()
		if err != nil {
			Errorf("close the log file %s fail, error is %s", LogFile.Name(), err)
		}
	}
	LogFile = pFile
}

func RemoveLogFile() {
	sPath, err := os.Getwd()
	if err != nil {
		Errorf("get the current file path error, the error is %s\n", err)
		return
	}
	sPath += "/" + LogDir
	fullPath, err := filepath.Abs(sPath)
	if err != nil {
		Errorf("get the absolute path error, the error is %s, the path is %s", err, fullPath)
		return
	}
	// 一天前
	d, err := time.ParseDuration("-24h")
	if err != nil {
		Errorf("parse the time error, the error is %s", err)
		return
	}
	cur := time.Now()
	curDate := cur.Add(d * 60).Format("20060102")

	filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Index(info.Name(), FileType) > 0 {
			nameArray := []byte(info.Name())
			var date []byte
			date = nameArray[len(FileName)+1:]
			sArray := strings.Split(string(date), ".")
			if IsMoreTwoMonth(sArray[0], curDate) { // 超过两个月删除
				err = os.Remove(sPath + info.Name())
				if err != nil {
					Errorf("delete the file %s error, error is ", info.Name(), err)
					return err
				}
			}
		}
		return nil
	})
}

func IsMoreTwoMonth(origDate, curDate string) bool {
	if strings.Compare(origDate, curDate) <= 0 {
		return true
	}
	return false
}
