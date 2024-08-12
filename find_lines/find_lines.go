package find_lines

import (
	"bufio"
	"github.com/GoHippo/slogpretty/sl"
	"io"
	"log/slog"
	"os"
	"sync"
)

type LineResult struct {
	Line      string
	PathFiles string
}

type FindLines struct {
	FindLinesOptions
	loader     chan string
	loaderSave chan LineResult
	wg         *sync.WaitGroup
	wgSave     *sync.WaitGroup
	arrResult  []LineResult
}
type FuncLineCheck func(scanner *bufio.Scanner) ([]string, bool, error)
type FuncFileCheck func(file []byte) ([]string, bool, error)

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
	scan := FindLines{FindLinesOptions: opt, wg: &sync.WaitGroup{}, wgSave: &sync.WaitGroup{}}

	if len(opt.PathFiles) == 0 {
		return nil, nil
	}

	scan.loader = make(chan string, len(opt.PathFiles))
	scan.loaderSave = make(chan LineResult, 1000000)
	scan.goSave()
	scan.goPool()

	for _, path := range opt.PathFiles {
		scan.wg.Add(1)
		scan.loader <- path
	}

	scan.wg.Wait()
	scan.wgSave.Wait()
	scan.close()

	return scan.arrResult, nil
}

func (srt *FindLines) action(path string) {

	file, err := os.OpenFile(path, os.O_RDONLY, 0444)
	if err != nil && srt.Log != nil {
		srt.Log.Error("Error opening file:"+path, sl.Err(err))
		return
	}
	defer file.Close()

	if srt.FuncCheckLine != nil {
		scanner := bufio.NewScanner(file)
		lines, ok, err := srt.FuncCheckLine(scanner)
		if err != nil {
			srt.Log.Error("Error checking lines in file", slog.String("path", path), sl.Err(err))
		}

		if ok {
			for _, line := range lines {
				srt.wgSave.Add(1)
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

}

func (srt *FindLines) goSave() {
	go func() {

		for {
			load := <-srt.loaderSave
			if load.PathFiles == "exit" {
				return
			}
			srt.arrResult = append(srt.arrResult, load)
			srt.wgSave.Done()
		}

	}()
}

func (srt *FindLines) goPool() {
	var countRunTreads int
	for _ = range srt.ThreadsCheckLines {
		countRunTreads++
		go func() {
			for {

				load := <-srt.loader

				if load == "exit" {
					return
				}

				srt.action(load)

				if srt.FuncSignalAdd != nil {
					srt.FuncSignalAdd(1)
				}

				srt.wg.Done()

			}

		}()
	}

	srt.Log.Debug("run threads find_lines:", slog.Int("threads", countRunTreads))
}

func (srt *FindLines) close() {
	defer close(srt.loader)
	defer close(srt.loaderSave)

	for _ = range srt.ThreadsCheckLines {
		srt.loader <- "exit"
	}
	srt.loaderSave <- LineResult{"", "exit"}
}
