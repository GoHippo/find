package find_lines

import "strings"

func FilteringRepeatLines(lines []LineResult, FuncSignalAdd func(i int)) (result []LineResult) {

	jar := make(map[string]byte)
	for _, line := range lines {
		line.Line = strings.TrimSpace(line.Line)
		if _, ok := jar[line.Line]; !ok {
			jar[line.Line] = 0
			result = append(result, line)
		}

		if FuncSignalAdd != nil {
			FuncSignalAdd(1)
		}
	}

	return result
}
