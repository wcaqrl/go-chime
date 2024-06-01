package pather

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func GetConfigFile(configFile, ePath string) (filePath string, err error) {
	if strings.HasPrefix(configFile, "/") {
		filePath = configFile
	} else if strings.HasPrefix(configFile, "./") {
		filePath = fmt.Sprintf("%s/%s", ePath, configFile[2:])
	} else {
		filePath = fmt.Sprintf("%s/%s", ePath, configFile)
	}
	// 检查配置文件是否存在
	if _, err = os.Stat(filePath); err != nil && !os.IsExist(err) {
		err = errors.New(fmt.Sprintf("config file %s not exists!", filePath))
	}
	return
}

func GetLogPath(ePath, rPath string) (logPath string) {
	if strings.HasPrefix(rPath, "/") {
		logPath = rPath
	} else if strings.HasPrefix(rPath, "./") {
		logPath = fmt.Sprintf("%s/%s", ePath, rPath[2:])
	} else {
		logPath = fmt.Sprintf("%s/%s", ePath, rPath)
	}
	return
}
