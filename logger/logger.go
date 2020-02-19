package logger

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rossmcdonald/telegram_hook"
	"github.com/sirupsen/logrus"
)

var (
	LOG *logrus.Logger
)

const (
	defaultLogDirName = "logs"   //日志目录
	maxLogFileSize    = 10 << 20 //最大日志文件大小10MB
	maxLogDirSize     = 1 << 30  //最大日志文件目录大小1GB
)

///清理空日志文件
func cleanEmptyLogFile(logDir string) {
	arr, err := ioutil.ReadDir(logDir)
	if nil != err {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ioutil.ReadDir error: %s\n", err)
		}
		return
	}
	for _, fi := range arr {
		if 0 == fi.Size() {
			absPath, err := filepath.Abs(filepath.Join(logDir, fi.Name()))
			if nil != err {
				fmt.Fprintf(os.Stderr, "filepath.Abs error: %s\n", err)
				continue
			}
			if err = os.Remove(absPath); nil != err {
				fmt.Fprintf(os.Stderr, "os.Remove %s error: %s\n", absPath, err)
				continue
			}
		}
	}
}

func InitLogger() {
	var once sync.Once

	if nil == LOG {
		once.Do(func() {
			//清理空文件
			logDir := os.Getenv("LOG_PATH")

			if logDir == "" {
				logDir = defaultLogDirName
			}
			cleanEmptyLogFile(logDir)

			//日志等级
			logLevel := logrus.DebugLevel

			//初始化日志对象
			LOG = logrus.New()
			LOG.SetLevel(logLevel)
			LOG.SetReportCaller(true)
			LOG.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})

			hook, err := telegram_hook.NewTelegramHook(
				"bitfinex-robot",
				os.Getenv("TELEGRAM_TOKEN"),
				os.Getenv("TELEGRAM_MANAGE_ID"),
				telegram_hook.WithAsync(true),
				telegram_hook.WithTimeout(30 * time.Second),
			)
			if err != nil {
				log.Panicf("Encountered error when creating Telegram hook: %s", err)
			}

			LOG.AddHook(hook)
		})
	}

}