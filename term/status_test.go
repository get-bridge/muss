package term

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteWithFixedStatusLine(t *testing.T) {
	var writer bytes.Buffer
	done := make(chan bool)
	outputCh := make(chan []byte, 10)
	statusCh := make(chan string, 1)

	ready := make(chan bool)
	go func() {
		WriteWithFixedStatusLine(&writer, outputCh, statusCh, done)
		close(ready)
	}()

	// pause to help the goroutine see ready channels in the right order
	pause := func() { time.Sleep(1 * time.Millisecond) }
	statusCh <- "ok"
	pause()
	outputCh <- []byte("log one")
	pause()
	statusCh <- "prefix\nyeah"
	pause()
	outputCh <- []byte("log two")
	pause()
	outputCh <- []byte("log three")
	pause()
	statusCh <- "done"
	pause()

	close(done)
	<-ready

	expOut := AnsiEraseToEnd +
		AnsiReset + "ok" + AnsiReset + AnsiStart +
		AnsiEraseToEnd + "log one\n" +
		AnsiReset + "ok" + AnsiReset + AnsiStart +
		AnsiEraseToEnd + AnsiReset + "prefix\nyeah" + AnsiReset + AnsiStart + AnsiUp +
		AnsiEraseToEnd + "log two\n" +
		AnsiReset + "prefix\nyeah" + AnsiReset + AnsiStart + AnsiUp +
		AnsiEraseToEnd + "log three\n" +
		AnsiReset + "prefix\nyeah" + AnsiReset + AnsiStart + AnsiUp +
		AnsiEraseToEnd + AnsiReset + "done" + AnsiReset + AnsiStart +
		AnsiEraseToEnd + AnsiReset + "" + AnsiReset + AnsiStart +
		""

	assert.Equal(t, expOut, writer.String())
}
