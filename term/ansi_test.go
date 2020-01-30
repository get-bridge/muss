package term

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stringToAnsiScanner(input string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(ScanLinesOrAnsiMovements)
	return scanner
}

func linesOrAnsiTokens(input string) []string {
	scanner := stringToAnsiScanner(input)
	tokens := make([]string, 0)
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil
	}
	return tokens
}

func TestScanANSI(t *testing.T) {
	assert.Equal(t, []string{"foo\r", "bar\r", "baz\n", ""},
		linesOrAnsiTokens("foo\rbar\rbaz\n"), "CR's and a NL")

	exp := []string{
		"Starting muss_work_1 ... \r\n",
		"Starting muss_ms_1   ... \r\n",
		"Starting muss_web_1  ... \r\n",
		"Starting muss_store_1 ... \r\n",
		"\x1B[4A",
		"\x1B[2K",
		"\r",
		"Starting muss_work_1  ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[4B",
		"\x1B[1A",
		"\x1B[2K",
		"\r",
		"Starting muss_store_1 ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[1B",
		"\x1B[2A",
		"\x1B[2K",
		"\r",
		"Starting muss_web_1   ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[2B",
		"\x1B[3A",
		"\x1B[2K",
		"\r",
		"Starting muss_ms_1    ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[3B",
		"Attaching to muss_work_1, muss_store_1, muss_web_1, muss_ms_1\n",
		"\x1B[36m", "ms_1     |\x1B[0m", " no tty\n",
		"\x1B[36m", "ms_1     |\x1B[0m", " # ms - 00001 #\n",
		"\x1B[33m", "store_1  |\x1B[0m", " no tty\n",
		"\x1B[33m", "store_1  |\x1B[0m", " # store - 00001 #\n",
		"\x1B[32m", "web_1    |\x1B[0m", " # web - 00001 #\r\n",
		"\x1B[35m", "work_1   |\x1B[0m", " # work - 00001 #\r",
		"Stopping muss_store_1 ... \r\n",
		"Stopping muss_ms_1    ... \r\n",
		"Stopping muss_work_1  ... \r\n",
		"Stopping muss_web_1   ... \r\n",
		"\x1B[2A",
		"\x1B[2K",
		"\r",
		"Stopping muss_work_1  ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[2B",
		"\x1B[4A",
		"\x1B[2K",
		"\r",
		"Stopping muss_store_1 ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[4B",
		"\x1B[3A",
		"\x1B[2K",
		"\r",
		"Stopping muss_ms_1    ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[3B",
		"\x1B[1A",
		"\x1B[2K",
		"\r",
		"Stopping muss_web_1   ... \x1B[32m",
		"done\x1B[0m",
		"\r",
		"\x1B[1B",
		"Gracefully stopping... (press Ctrl+C again to force)\n",
		"",
	}
	assert.Equal(t, exp, linesOrAnsiTokens(strings.Join(exp, "")))
}
