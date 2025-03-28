package complex_detection

import "io"

// Reader - интерфейс для чтения данных
// want "identifier .Reader. is only used in test files but is not part of test files"
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer - интерфейс для записи данных
// want "identifier .Writer. is only used in test files but is not part of test files"
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer - интерфейс для закрытия ресурсов
// want "identifier .Closer. is only used in test files but is not part of test files"
type Closer interface {
	Close() error
}

// ReadWriter - композитный интерфейс, объединяющий Reader и Writer
// want "identifier .ReadWriter. is only used in test files but is not part of test files"
type ReadWriter interface {
	Reader
	Writer
}

// ReadWriteCloser - композитный интерфейс, объединяющий Reader, Writer и Closer
// want "identifier .ReadWriteCloser. is only used in test files but is not part of test files"
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// CustomReader - конкретная реализация интерфейса Reader
// want "identifier .CustomReader. is only used in test files but is not part of test files"
type CustomReader struct {
	data []byte
	pos  int
}

// Read реализует интерфейс Reader
// want "identifier .Read. is only used in test files but is not part of test files"
func (r *CustomReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// CustomWriter - конкретная реализация интерфейса Writer
// want "identifier .CustomWriter. is only used in test files but is not part of test files"
type CustomWriter struct {
	data []byte
}

// Write реализует интерфейс Writer
// want "identifier .Write. is only used in test files but is not part of test files"
func (w *CustomWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

// CustomReadWriter реализует композитный интерфейс ReadWriter
// want "identifier .CustomReadWriter. is only used in test files but is not part of test files"
type CustomReadWriter struct {
	CustomReader
	CustomWriter
}

// FullImplementation реализует все три интерфейса
// want "identifier .FullImplementation. is only used in test files but is not part of test files"
type FullImplementation struct {
	CustomReadWriter
}

// Close реализует интерфейс Closer
// want "identifier .Close. is only used in test files but is not part of test files"
func (f *FullImplementation) Close() error {
	// Закрываем ресурсы
	return nil
}

// Process обрабатывает данные через интерфейс ReadWriter
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

// ProcessAndClose обрабатывает данные и закрывает ресурс
// want "identifier .ProcessAndClose. is only used in test files but is not part of test files"
func ProcessAndClose(rwc ReadWriteCloser, data []byte) ([]byte, error) {
	defer rwc.Close()
	return Process(rwc, data)
}
