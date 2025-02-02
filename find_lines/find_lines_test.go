package find_lines

import (
	"bytes"
	"github.com/GoHippo/find/find_pathes"
	"github.com/GoHippo/slogpretty/slogpretty"
	"os"
	"slices"
	"testing"
)

func TestFindLinesStart(t *testing.T) {
	
	if err := create_file_for_test(); err != nil {
		t.Error("Create test.txt file error: " + err.Error())
	}
	defer os.Remove("./test.txt")
	
	fCheck := func(res []byte) ([]byte, bool, error) {
		line := string(bytes.TrimSpace(res))
		
		if line == "test_lines" {
			return []byte(line), true, nil
		}
		return []byte(line), false, nil
	}
	
	log := slogpretty.SetupPrettySlog()
	
	arr, err := NewFindLines(FindLinesOptions{
		FuncCheck: fCheck,
		Log:       log,
		FindOptions: find_pathes.FindOption{
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
	
	if !slices.Equal(arr[0].Line, []byte("test_lines")) && !slices.Equal(arr[1].Line, []byte("test_lines")) {
		t.Log(len(arr), string(arr[0].Line), string(arr[1].Line))
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
