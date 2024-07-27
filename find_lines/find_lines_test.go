package find_lines

import (
	"github.com/GoHippo/find"
	"github.com/GoHippo/slogpretty/slogpretty"
	"os"
	"strings"
	"testing"
)

func TestFindLinesStart(t *testing.T) {
	
	if err := create_file_for_test(); err != nil {
		t.Error("Create test.txt file error: " + err.Error())
	}
	defer os.Remove("./test.txt")
	
	fCheck := func(line string) (string, bool, error) {
		line = strings.TrimSpace(line)
		if line == "test_lines" {
			return line, true, nil
		}
		return "", false, nil
	}
	
	log := slogpretty.SetupPrettySlog()
	
	arr, err := FindLinesStart(FindLinesOptions{
		LineCheck: fCheck,
		Log:       log,
		FindOptions: find.FindOption{
			Log:         log,
			FindName:    ".txt",
			Path:        "./",
			IsFile:      true,
			Threads:     10,
			SignalFind:  nil,
			MaxSizeFile: 0,
		},
		
		ThreadsCheckLines: 10,
		SignalBar:         nil,
	})
	
	if err != nil {
		t.Fatal(err)
	}
	
	if arr[0].Line != "test_lines" && arr[3].Line != "test_lines" {
		t.Fatal("Error TestFindLinesStart ")
	}
}

func create_file_for_test() error {
	
	data := []byte("test_lines\n1-test_lines\n2-test_lines\ntest_lines")
	
	if err := os.WriteFile("./test.txt", data, 0755); err != nil {
		return err
	}
	
	return nil
	
}
