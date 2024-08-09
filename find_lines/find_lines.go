package find_lines

import (
	"bufio"
	"github.com/GoHippo/slogpretty/sl"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

type LineResult struct {
	Line      []byte
	PathFiles string
}

type FindLines struct {
	FindLinesOptions
	loader     chan string
	loaderSave chan LineResult
	wg         *sync.WaitGroup
	arrResult  []LineResult
}
type FuncLineCheck func(line []byte) ([]byte, bool, error)
type FuncFileCheck func(file []byte) ([][]byte, bool, error)

// SignalBar и Log не обязательны
type FindLinesOptions struct {
	PathFiles     []string
	FuncCheckLine FuncLineCheck
	FuncCheckFile FuncFileCheck
	Log           *slog.Logger

	ThreadsCheckLines int
	FuncSignalAdd     func(i int)
}

func NewFindLines(opt FindLinesOptions) ([]LineResult, error) {
	scan := FindLines{FindLinesOptions: opt, wg: &sync.WaitGroup{}}

	if len(opt.PathFiles) == 0 {
		return nil, nil
	}

	scan.loader = make(chan string, len(opt.PathFiles))
	scan.loaderSave = make(chan LineResult, 1000)
	scan.goSave()
	scan.goPool()

	for _, path := range opt.PathFiles {
		scan.wg.Add(1)
		scan.loader <- path
	}

	scan.wg.Wait()
	scan.close()

	return scan.arrResult, nil
}

func (srt *FindLines) action(path string) {

	file, err := os.OpenFile(path, os.O_RDONLY, 0444)
	if err != nil && srt.Log != nil {
		srt.Log.Error("Error opening file:"+path, sl.Err(err))
		return
	}

	if srt.FuncCheckLine != nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {

			line, ok, err := srt.FuncCheckLine(scanner.Bytes())

			if err != nil && srt.Log != nil {
				srt.Log.Error("Error checking line in file:"+path, sl.Err(err))
				continue
			}

			if ok {
				srt.loaderSave <- LineResult{line, path}
			}

		}
	}

	if srt.FuncCheckFile != nil {
		data, err := io.ReadAll(file)
		if err != nil && srt.Log != nil {
			srt.Log.Error("Error read file:"+path, sl.Err(err))
			return
		}

		lines, ok, err := srt.FuncCheckFile(data)

		if err != nil && srt.Log != nil {
			srt.Log.Error("Error checking file:"+path, sl.Err(err))
			return
		}

		if ok {
			for _, line := range lines {
				srt.loaderSave <- LineResult{line, path}
			}
		}

	}

	file.Close()
}

func (srt *FindLines) goSave() {
	go func() {
		for {
			select {
			case load := <-srt.loaderSave:
				if load.PathFiles == "exit" {
					return
				}

				srt.arrResult = append(srt.arrResult, load)

			default:
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()
}

func (srt *FindLines) goPool() {
	for _ = range srt.ThreadsCheckLines {
		go func() {
			for {
				select {
				case load := <-srt.loader:

					if load == "exit" {
						return
					}

					srt.action(load)

					if srt.FuncSignalAdd != nil {
						srt.FuncSignalAdd(1)
					}

					srt.wg.Done()

				default:
					time.Sleep(time.Millisecond * 10)
				}
			}

		}()
	}
}

func (srt *FindLines) close() {
	defer close(srt.loader)
	defer close(srt.loaderSave)

	for _ = range srt.ThreadsCheckLines {
		srt.loader <- "exit"
	}
	srt.loaderSave <- LineResult{[]byte{}, "exit"}
}
