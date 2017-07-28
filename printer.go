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
	QUICKPAY  = "quickpay"
	PHANTOM   = "phantom"
	ANGRYCARD = "angrycard"
	BIGCAT    = "bigcat"
	MAGISYNC  = "magisync"
	MAGISETT  = "magisett"
	SHOPCODE  = "shopcode"
	MEDUSA    = "medusa"
	GOVERNOR  = "governor"
	FileType  = ".log"
)

var progArr = []string{
	QUICKPAY,
	PHANTOM,
	ANGRYCARD,
	BIGCAT,
	MAGISYNC,
	MAGISETT,
	SHOPCODE,
	MEDUSA,
	GOVERNOR,
}

var LogDir = "logs/"
var LogFile *os.File = nil
var szFileName string = ""
var bNeedScrolling = false

func init() {
	// 日志初始化
	PreLog()
	// RemoveLogFile()
	c := cron.New()
	c.AddFunc("0 0 0 * * ?", PreLog) // 每隔一天
	// c.AddFunc("0/3 * * * * ?", PreLog) // 每隔3秒
	// c.AddFunc("0 0 0 1 * *", RemoveLogFile) // 每隔一个月执行
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

func SetFileName(fileName string) {
	szFileName = fileName
	bNeedScrolling = needScrolling(fileName)
}

func needScrolling(fileName string) bool {
	for _, progName := range progArr {
		if strings.Contains(fileName, progName) {
			return true
		}
	}
	return false
}

// 由于该方法是定时任务，第一次和程序一起启动，如果失败，在控制台输出错误信息，如果后续定时任务出现处理失败，则沿用原来的log文件
func PreLog() {
	Infof("LogDir: %s\n", LogDir)
	if szFileName == "" {
		_, szFileName = filepath.Split(os.Args[0])
		bNeedScrolling = needScrolling(szFileName)
		return
	}

	if !bNeedScrolling {
		return
	}

	currLogFileName := LogDir + szFileName + FileType
	// 类似phantom.log.20160512与quickpay.log.20160512
	sName := LogDir + szFileName + FileType + "." + time.Now().AddDate(0, 0, -1).Format("20060102")
	err := os.Rename(currLogFileName, sName) // 重命名当前log文件名
	if err != nil {
		slog := fmt.Sprintf("rename the log file error, the error is %s\n", err)
		if LogFile == nil {
			fmt.Println(slog)
			return
		}
		Error(slog)
		return
	}

	pFile, err := os.Create(currLogFileName)
	if (err != nil) || (pFile == nil) {
		slog := fmt.Sprintf("create the log file error, the error is %s\n", err)
		if LogFile == nil {
			fmt.Println(slog)
			return
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
		if (info != nil) && (!info.IsDir()) && (strings.Index(info.Name(), FileType) > 0) {
			nameArray := []byte(info.Name())
			var date []byte
			date = nameArray[len(szFileName)+1:]
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

// true:超过两个月
func IsMoreTwoMonth(origDate, curDate string) bool {
	if strings.Compare(origDate, curDate) <= 0 {
		return true
	}
	return false
}
