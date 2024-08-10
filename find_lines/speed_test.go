package find_lines

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GoHippo/find/find_pathes"
	"github.com/GoHippo/find/pkg/pterm_tools/pb_default"
	"github.com/GoHippo/find/pkg/pterm_tools/pb_spinner"
	"github.com/GoHippo/slogpretty/slogpretty"
	"github.com/schollz/progressbar/v3"
	log2 "log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var log = slogpretty.SetupPrettySlog(slog.LevelInfo)

func TestSpeedNewFindLines(t *testing.T) {

	go func() {
		log2.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	path := `/home/meteoroot/ShareVM/WAGNER TRAFFIC #17 Thanks For Subscription`

	arrPath := find(path, log)

	arrLines, err := parse(arrPath, log)
	if err != nil {
		t.Error(err)
		return
	}

	log.Info("Count lines:" + strconv.Itoa(len(arrLines)))

}

func find(path string, log *slog.Logger) []string {
	pb := pb_spinner.NewSpinnerBar("Scan folder")
	arrPaths := find_pathes.NewFindPath(find_pathes.FindOption{
		Log:           log,
		FindName:      ".txt",
		Path:          path,
		IsFile:        true,
		Threads:       30,
		FuncSignalAdd: pb.Add,
		MaxSizeFile:   0,
	})
	pb.Close("files find")

	return arrPaths
}

func parse(arrPaths []string, log *slog.Logger) ([]LineResult, error) {
	// ====================== parse ======================
	fmt.Println("")
	log.Info("The beginning of token parsing in regex mode")

	pbLines := pb_default.NewPB(len(arrPaths), "Parse token")

	arrFindLines, err := NewFindLines(FindLinesOptions{
		PathFiles:         arrPaths,
		FuncCheckLine:     actionLines(),
		FuncCheckFile:     nil,
		Log:               log,
		ThreadsCheckLines: 30,
		FuncSignalAdd:     pbLines.Add,
	})
	//pb.Close("token")

	return arrFindLines, err
}

// ====================== Msg ======================

type pbv3 struct {
	*progressbar.ProgressBar
}

func newPBV3(max int) pbv3 {
	return pbv3{progressbar.Default(int64(max), "parse tokens")}
}

func (pb pbv3) Add(i int) {
	pb.Add64(int64(i))
}

// ====================== ActionBox ======================

func actionLines() func(scanner *bufio.Scanner) ([]string, bool, error) {

	return func(scanner *bufio.Scanner) ([]string, bool, error) {
		var err error
		var arrTokens []string

		// ====================== Regex ======================
		re1 := regexp.MustCompile(`1//.{100}`)
		re2 := regexp.MustCompile(`\d{21}`)

		reLumma := regexp.MustCompile(`^[A-Za-z0-9_-]{235}$`)

		for scanner.Scan() {
			line := scanner.Bytes()

			if err != nil {
				continue
			}

			for _, str := range bytes.Split(line, []byte("\n")) {
				str = bytes.TrimSpace(str)
				// Проверяем наличие обоих условий
				if re1.Match(str) && re2.Match(str) {

					var token string
					if matches := re1.Find(str); len(matches) > 0 {
						token = string(matches)
					}

					if matches := re2.Find(str); len(matches) > 0 {
						token = token + ":" + string(matches)
					}

					arrTokens = append(arrTokens, token)
				}

				if reLumma.Match(str) {
					if matches := reLumma.Find(str); len(matches) > 0 {
						arrTokens = append(arrTokens, string(matches))
					}
				}
			}
		}

		return arrTokens, len(arrTokens) != 0, err
	}
}

func actionFile() func(file []byte) ([]string, bool, error) {
	return func(file []byte) ([]string, bool, error) {
		f := FileFormatTokens{}
		if arr := f.CheckFile(file, log); len(arr) > 0 {
			return arr, true, nil
		}
		return nil, false, nil
	}
}

type FileFormatTokens struct {
	List []struct {
		Service string `json:"service"`
		Token   string `json:"token"`
	} `json:"list"`
}

func (f *FileFormatTokens) CheckFile(data []byte, log *slog.Logger) (arr []string) {

	if indexJsonStart := bytes.IndexByte(data, byte('{')); indexJsonStart > 0 {
		data = data[indexJsonStart:]
	}

	err := json.Unmarshal(data, f)
	if err != nil {
		return nil
	}

	if len(f.List) == 0 {
		return nil
	}

	for _, l := range f.List {

		if strings.Contains(l.Service, "fake") {
			continue
		}

		sp := strings.Split(l.Service, "-")
		if len(sp) != 2 {
			log.Error("json format token err in service", slog.String("service", l.Service))
			continue
		}

		if len(sp[1]) != 21 {
			continue
		}

		arr = append(arr, fmt.Sprintf("%v:%v", l.Token, sp[1]))
	}
	return arr

}
