package tracker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
)

func NewLogTracker(loggers []Logger) (*LogTracker, error) {
	tracker := &LogTracker{
		loggers: make(map[LoggerKey]Logger),
	}
	for i, val := range loggers {
		key := &LoggerKey{}
		if val.Name == "" {
			return tracker, errors.New("Logger name was empty, each logger tracker requires a name")
		}
		key.Name = val.Name

		if val.Id == "" {
			key.Id = fmt.Sprintf("%d", i)
		} else {
			key.Id = val.Id
		}
		logger := Logger{
			LoggerKey: key,
			Logger:    val.Logger,
		}
		tracker.loggers[*key] = logger
	}

	return tracker, nil
}

type LoggerKey struct {
	Id, Name string
}

type Logger struct {
	*LoggerKey
	Logger io.ReadCloser
}

type LogTracker struct {
	loggers map[LoggerKey]Logger
}

type StopFunc func()

func (l *LogTracker) Start() StopFunc {
	writerContext, cancel := context.WithCancel(context.Background())
	go l.writerCollector(writerContext)
	return func() { cancel() }
}

func (l *LogTracker) AddLogReader(logger io.ReadCloser, name, Id string) error {
	if Id == "" || name == "" {
		return errors.New("Either Id or name is empty, both must be populated")
	}
	key := &LoggerKey{
		Id:   Id,
		Name: name,
	}
	if _, ok := l.loggers[*key]; ok {
		return fmt.Errorf("Logger with key (%+v) already exists", key)
	}
	l.loggers[*key] = Logger{
		LoggerKey: key,
		Logger:    logger,
	}
	return nil
}

func (l *LogTracker) writerCollector(ctx context.Context) {
	logChan := make(chan string)
	for _, val := range l.loggers {
		go writer(val.Logger, logChan)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case val := <-logChan:
			fmt.Println(val)
		}
	}
}

func writer(stdout io.ReadCloser, logChan chan string) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		logChan <- scanner.Text()
	}
}
