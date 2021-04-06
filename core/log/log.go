package log

import "github.com/idena-network/idena-indexer/log"

func GetLogFileHandler(fileName string, logFileSize int) (log.Handler, error) {
	fileHandler, _ := log.RotatingFileHandler(fileName, uint(logFileSize*1024), log.TerminalFormat(false))
	return fileHandler, nil
}

func NewFileLogger(fileName string, fileSize int) (log.Logger, error) {
	fileHandler, err := GetLogFileHandler(fileName, fileSize)
	if err != nil {
		return nil, err
	}
	l := log.New()
	l.SetHandler(fileHandler)
	return l, nil
}
