package complex_detection

import "io"

// Reader - interface for reading data
// want "identifier .Reader. is only used in test files but is not part of test files"
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer - interface for writing data
// want "identifier .Writer. is only used in test files but is not part of test files"
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer - interface for closing resources
// want "identifier .Closer. is only used in test files but is not part of test files"
type Closer interface {
	Close() error
}

// ReadWriter - composite interface combining Reader and Writer
// want "identifier .ReadWriter. is only used in test files but is not part of test files"
type ReadWriter interface {
	Reader
	Writer
}

// ReadWriteCloser - composite interface combining Reader, Writer and Closer
// want "identifier .ReadWriteCloser. is only used in test files but is not part of test files"
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// CustomReader - concrete implementation of the Reader interface
// want "identifier .CustomReader. is only used in test files but is not part of test files"
type CustomReader struct {
	data []byte
	pos  int
}

// Read implements the Reader interface
// want "identifier .Read. is only used in test files but is not part of test files"
func (r *CustomReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// CustomWriter - concrete implementation of the Writer interface
// want "identifier .CustomWriter. is only used in test files but is not part of test files"
type CustomWriter struct {
	data []byte
}

// Write implements the Writer interface
// want "identifier .Write. is only used in test files but is not part of test files"
func (w *CustomWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

// CustomReadWriter implements the composite ReadWriter interface
// want "identifier .CustomReadWriter. is only used in test files but is not part of test files"
type CustomReadWriter struct {
	CustomReader
	CustomWriter
}

// FullImplementation implements all three interfaces
// want "identifier .FullImplementation. is only used in test files but is not part of test files"
type FullImplementation struct {
	CustomReadWriter
}

// Close implements the Closer interface
// want "identifier .Close. is only used in test files but is not part of test files"
func (f *FullImplementation) Close() error {
	// Close resources
	return nil
}

// Process processes data through the ReadWriter interface
// want "identifier .Process. is only used in test files but is not part of test files"
func Process(rw ReadWriter, data []byte) ([]byte, error) {
	_, err := rw.Write(data)
	if err != nil {
		return nil, err
	}

	result := make([]byte, len(data))
	_, err = rw.Read(result)
	return result, err
}

// ProcessAndClose processes data and closes the resource
// want "identifier .ProcessAndClose. is only used in test files but is not part of test files"
func ProcessAndClose(rwc ReadWriteCloser, data []byte) ([]byte, error) {
	defer rwc.Close()
	return Process(rwc, data)
}
