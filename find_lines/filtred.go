package find_lines

import "strings"

func FilteringRepeatLines(lines []string, FuncSignalAdd func(i int)) (result []string) {

	jar := make(map[string]byte)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if _, ok := jar[line]; !ok {
			jar[line] = 0
			result = append(result, line)
		}

		if FuncSignalAdd != nil {
			FuncSignalAdd(1)
		}
	}

	return result
}
