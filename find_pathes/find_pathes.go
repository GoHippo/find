package find_pathes

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FindScan struct {
	FindOption
	loaderScan     chan string
	savePathLoader chan string
	wg             *sync.WaitGroup
	wgSave         *sync.WaitGroup
	arrPathCookies []string
	signalExit     chan struct{}
	signalSaveExit chan struct{}
}

type FindOption struct {
	Log           *slog.Logger
	FindName      string
	Path          string
	IsFile        bool
	Threads       int
	FuncSignalAdd func(i int)
	MaxSizeFile   int64
}

func NewFindPath(opt FindOption) []string {
	opt.FindName = strings.ToLower(opt.FindName)
	cfs := FindScan{
		FindOption:     opt,
		loaderScan:     make(chan string, 20000000), // заблокируется если переполнить буфер
		savePathLoader: make(chan string),
		wg:             &sync.WaitGroup{},
		wgSave:         &sync.WaitGroup{},
		arrPathCookies: make([]string, 0),
		signalExit:     make(chan struct{}),
		signalSaveExit: make(chan struct{}),
	}

	cfs.wg.Add(1)
	cfs.loaderScan <- opt.Path
	cfs.goHandleSavePath()
	cfs.goPool()
	cfs.wg.Wait()
	cfs.wgSave.Wait()

	cfs.close()

	return cfs.arrPathCookies
}

func (fs *FindScan) goPool() {
	for _ = range fs.Threads {
		go func() {
			for {
				select {
				case p := <-fs.loaderScan:
					fs.scan(p)
				case _ = <-fs.signalExit:
					return
				default:
					time.Sleep(time.Millisecond * 30)
				}
			}
		}()
		time.Sleep(time.Millisecond * 10)
	}
}

func (fs *FindScan) scan(p string) {
	defer fs.wg.Done()
	// defer bar.Add(1)
	dir, err := os.ReadDir(p)
	if err != nil {
		fs.Log.Error(err.Error())
		return
	}
	for _, sc := range dir {

		switch {
		case sc.IsDir():
			scPath := filepath.Join(p, sc.Name())
			if !fs.IsFile && strings.Contains(strings.ToLower(sc.Name()), fs.FindName) {
				fs.wgSave.Add(1)
				fs.savePathLoader <- scPath
			}
			fs.wg.Add(1)
			fs.loaderScan <- scPath

		case !sc.IsDir() && fs.IsFile && strings.Contains(strings.ToLower(sc.Name()), fs.FindName):

			if fs.MaxSizeFile != 0 {
				fileInfo, err := sc.Info()
				if err != nil {
					fs.Log.Error("Get File Info:" + err.Error())
					continue
				}
				if fileInfo.Size() > fs.MaxSizeFile {
					continue
				}
			}

			fs.wgSave.Add(1)
			fs.savePathLoader <- filepath.ToSlash(filepath.Join(p, sc.Name()))
		}
	}
}

func (fs *FindScan) goHandleSavePath() {
	go func() {
		for {
			select {
			case p := <-fs.savePathLoader:
				fs.arrPathCookies = append(fs.arrPathCookies, p)
				if fs.FuncSignalAdd != nil {
					fs.FuncSignalAdd(1)
				}
				fs.wgSave.Done()
			case _ = <-fs.signalSaveExit:
				return
			default:
				time.Sleep(time.Millisecond)
			}
		}
	}()
}

func (fs *FindScan) close() {
	for _ = range fs.Threads {
		fs.signalExit <- struct{}{}
	}
	fs.signalSaveExit <- struct{}{}

	close(fs.savePathLoader)
	close(fs.signalExit)
	close(fs.loaderScan)
	close(fs.signalSaveExit)
}
