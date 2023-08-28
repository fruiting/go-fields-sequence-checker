package main

import (
	"fmt"
	"os"
	"sync"
)

const (
	// "\t" char byte
	tabSpecialCharByte uint8 = 9
	// "\n" char byte
	lineBreakByte uint8 = 10
	// "{" char byte
	openBraceByte uint8 = 123
	// "}" char byte
	closeBraceByte uint8 = 125
)

var parseFlag = []byte("`parse: \"true\"`")

func main() {
	os.Setenv("FILE", "/Users/romanspirin/go/src/self-projects/learn-project/main.go")
	os.Setenv("STRUCT", "SomeStruct")
	path := os.Getenv("FILE")
	if path == "" {
		fmt.Println("FILE is required")

		return
	}

	structVal := os.Getenv("STRUCT")
	if structVal == "" {
		fmt.Println("STRUCT is required")

		return
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(fmt.Errorf("can't read file: %w", err))

		return
	}

	compareBytes := []byte(fmt.Sprintf("type %s struct {", structVal))
	parsedStruct := findStructInCode(bytes, compareBytes)
	if parsedStruct == nil {
		fmt.Println(fmt.Sprintf("error: struct %s not found", structVal))

		return
	}

	parsedStruct = []byte("type SomeStruct struct {\n\tval       string`parse: \"true\"`\n\tvalInt    int`parse: \"true\"`\n\tvalBool   bool`parse: \"true\"`\n\tstructVal struct{}`parse: \"true\"`\n}")
	parse := parseStruct(parsedStruct)

	//wg := &sync.WaitGroup{}

	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	m := make(map[string][][]byte, 0)
	for _, line := range parse {
		wg.Add(1)
		go func(line []byte) {
			defer wg.Done()

			mu.Lock()
			defer mu.Unlock()

			preparedLine, fieldType := prepareLine(line)
			m[fieldType] = append(m[fieldType], preparedLine)
		}(line)
	}

	wg.Wait()

	fmt.Println(parse)
}

func prepareLine(line []byte) ([]byte, string) {
	var start, typeStart int
	for i := 0; i < len(line); i++ {
		if line[i] == lineBreakByte || line[i] == tabSpecialCharByte {
			continue
		}
		if start == 0 {
			start = i
		}
		s := string(line[i])
		fmt.Println(s)
		if string(line[i]) == " " {
			typeStart = i + 1
		}

		if string(line[i]) == "`" {
			readyLine := make([]byte, i-2, i-2)
			fieldType := make([]byte, i-typeStart, i-typeStart)
			copy(readyLine, line[start:i])
			copy(fieldType, line[typeStart:i])

			return readyLine, string(fieldType)
		}
	}

	return nil, ""
}

func parseStruct(structToParse []byte) [][]byte {
	var offset int
	parse := make([][]byte, 0)
	for i := 0; i < len(structToParse); i++ {
		if structToParse[i] != lineBreakByte {
			continue
		}

		parseFlagStrCounter := len(parseFlag) - 1
		line := structToParse[offset:i]
		var canBeParsed bool

		for j := len(line) - 1; j >= 0; j-- {
			if line[j] != parseFlag[parseFlagStrCounter] {
				break
			}

			if parseFlagStrCounter == 0 {
				canBeParsed = true
				break
			}
			parseFlagStrCounter--
		}

		if canBeParsed {
			parse = append(parse, line)
		}

		offset = i
	}

	return parse
}

func findStructInCode(b, compareBytes []byte) []byte {
	stringBytes := make([]byte, len(compareBytes), len(compareBytes))
	var bracketsStack []string // stack for checking inserted structs
	var foundStruct bool
	var offset int

	for i := 0; i < len(b); i++ {
		if b[i] != lineBreakByte {
			continue
		}

		// checking if specific structure found in file, and it's parsing is in progress
		if foundStruct {
			for j := offset; j < i; j++ {
				if b[j] == openBraceByte {
					bracketsStack = append(bracketsStack, string(openBraceByte))
				} else if b[j] == closeBraceByte {
					bracketsStack = bracketsStack[:len(bracketsStack)-1]
				}
			}
			// if found the end of struct
			if b[i-1] == closeBraceByte && len(bracketsStack) == 0 {
				stringBytes = append(stringBytes, b[offset:i]...)

				return stringBytes
			}

			if b[offset] != lineBreakByte {
				stringBytes = append(stringBytes, lineBreakByte)
			}
			stringBytes = append(stringBytes, b[offset:i]...)
			offset = i
		} else {
			// checking if line len is less that needed for struct init
			if len(b[offset:i]) < len(compareBytes) {
				offset = i + 1
				continue
			}

			// for using less cap
			copy(stringBytes, b[offset:i])

			// comparing strings
			foundStruct = true
			for j := 0; j < len(stringBytes); j++ {
				if stringBytes[j] != compareBytes[j] && stringBytes[j] == closeBraceByte {
					return stringBytes
				} else if stringBytes[j] != compareBytes[j] {
					foundStruct = false
				}
			}

			if stringBytes[len(stringBytes)-1] == openBraceByte {
				bracketsStack = append(bracketsStack, string(openBraceByte))
			}

			offset = i + 1
		}
	}

	return nil
}
