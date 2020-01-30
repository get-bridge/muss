package term

import (
	"bufio"
)

// ANSI Escape Sequences
const (
	AnsiUp         = "\033[0A"
	AnsiStart      = "\033[0G"
	AnsiEraseToEnd = "\033[0J"
	AnsiReset      = "\033[0m"
	// AnsiDown       = "\033[0B"
	// AnsiEraseLine  = "\033[2K"

	CR  = '\r'
	LF  = '\n'
	Esc = 0x1b
	la  = 'a'
	lz  = 'z'
	lA  = 'A'
	lZ  = 'Z'
)

// ScanLinesOrAnsiMovements is similar to ScanLines except that it will return
// tokens that end at ansi movements or newlines. Tokens will contain the
// terminator so that they can be printed directly.
func ScanLinesOrAnsiMovements(data []byte, atEOF bool) (advance int, token []byte, err error) {
	ansiSeq := false
	next := byte(0)
	this := byte(0)
	length := len(data)
	beforeLast := length - 1
	send := false

	for i := 0; i < length; i++ {
		this = data[i]
		if i < beforeLast {
			next = data[i+1]
		} else {
			next = byte(0)
		}
		send = false

		if this == Esc && next == '[' {
			ansiSeq = true
		} else if ansiSeq {
			// If inside an ansi sequence and it is ending, return the text so far.
			if (data[i] >= la && data[i] <= lz) || (data[i] >= lA && data[i] <= lZ) {
				send = true
			}
		} else if data[i] == CR && next != LF {
			// Send CR unless it's a crlf then wait and send both.
			send = true
		} else if data[i] == LF {
			send = true
		}
		if send {
			// Include final byte in token and signal to advance beyond it.
			return i + 1, data[:i+1], nil
		}
	}

	if !atEOF {
		return 0, nil, nil
	}

	// There is one final token to be delivered, which may be the empty string.
	// Returning bufio.ErrFinalToken here tells Scan there are no more tokens after this
	// but does not trigger an error to be returned from Scan itself.
	return 0, data, bufio.ErrFinalToken
}
