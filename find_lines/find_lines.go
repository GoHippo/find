package find_lines

import (
	"bufio"
	"github.com/GoHippo/find"
	"github.com/GoHippo/slogpretty/sl"
	"log/slog"
	"os"
	"sync"
	"time"
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
	arrResult  []LineResult
}

type FindLinesOptions struct {
	LineCheck   func(line string) (string, bool, error)
	Log         *slog.Logger
	FindOptions find.FindOption
	
	ThreadsCheckLines int
	SignalBar         chan int
}

func FindLinesStart(opt FindLinesOptions) ([]LineResult, error) {
	scan := FindLines{FindLinesOptions: opt, wg: &sync.WaitGroup{}}
	
	arrPathFiles := find.NewFindPath(opt.FindOptions)
	if len(arrPathFiles) == 0 {
		return nil, nil
	}
	
	if opt.Log != nil {
		opt.Log.Info("The beginning of files parsing.")
	}
	
	scan.loader = make(chan string, len(arrPathFiles))
	scan.loaderSave = make(chan LineResult, 1000)
	scan.goSave()
	scan.goPool()
	
	for _, path := range arrPathFiles {
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
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		
		line, ok, err := srt.LineCheck(scanner.Text())
		
		if err != nil && srt.Log != nil {
			srt.Log.Error("Error checking file:"+path, sl.Err(err))
			continue
		}
		
		if ok {
			srt.loaderSave <- LineResult{line, path}
		}
		
	}
	
	file.Close()
}

func (srt *FindLines) goSave() {
	go func() {
		for {
			select {
			case load := <-srt.loaderSave:
				if load.Line == "exit" {
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
					
					if srt.SignalBar != nil {
						srt.SignalBar <- 1
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
	srt.loaderSave <- LineResult{"exit", ""}
}
