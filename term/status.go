package term

import (
	"io"
	"strings"
)

func writeWithStatusLine(out io.Writer, output []byte, status string) {
	// Cursor should already be where we want to start writing.
	out.Write([]byte(AnsiEraseToEnd))

	if len(output) > 0 {
		out.Write(output)       // print (trimmed) output
		out.Write([]byte("\n")) // end it
	}

	// Wrap status in resets to avoid bleeding and put cursor back at start.
	out.Write([]byte(AnsiReset + status + AnsiReset + AnsiStart))

	// If the status consists of more than one line
	// move the cursor up to the first line
	// to be in position for the next write.
	statusLines := strings.Count(status, "\n") + 1
	if statusLines > 1 {
		cursorUp := []byte{}
		for i := 1; i < statusLines; i++ {
			cursorUp = append(cursorUp, []byte(AnsiUp)...)
		}
		out.Write(cursorUp)
	}
}

// WriteWithFixedStatusLine takes a writer, a channel for []byte to write to it,
// a channel for status updates to always be at the bottom, and a done chan to
// exit the goroutine.
func WriteWithFixedStatusLine(writer io.Writer, outputCh chan []byte, statusCh chan string, done chan bool) {
	var status string
	for {
		select {
		case update := <-statusCh:
			status = update // update the "global"
			writeWithStatusLine(writer, nil, status)
		case output := <-outputCh:
			writeWithStatusLine(writer, output, status)
		case <-done:
			writeWithStatusLine(writer, nil, "")
			return
		}
	}
}
