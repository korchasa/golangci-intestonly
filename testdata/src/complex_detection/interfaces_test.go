package complex_detection

import (
	"testing"
)

func TestSimpleInterfaces(t *testing.T) {
	// Создаем реализацию Reader
	reader := &CustomReader{
		data: []byte("test data for reader"),
	}

	// Проверяем реализацию Reader
	buf := make([]byte, 4)
	n, err := reader.Read(buf)
	if err != nil {
		t.Errorf("Error reading: %v", err)
	}
	t.Logf("Read %d bytes: %s", n, buf)

	// Создаем реализацию Writer
	writer := &CustomWriter{}

	// Проверяем реализацию Writer
	n, err = writer.Write([]byte("test"))
	if err != nil {
		t.Errorf("Error writing: %v", err)
	}
	t.Logf("Wrote %d bytes, buffer now: %s", n, writer.data)
}

func TestCompositeInterfaces(t *testing.T) {
	// Создаем реализацию композитного интерфейса ReadWriter
	readWriter := &CustomReadWriter{
		CustomReader: CustomReader{
			data: []byte("readwriter test"),
		},
	}

	// Используем через интерфейс ReadWriter
	var rw ReadWriter = readWriter

	// Записываем данные
	testData := []byte("interface test")
	_, err := rw.Write(testData)
	if err != nil {
		t.Errorf("Error writing through interface: %v", err)
	}

	// Читаем данные
	buf := make([]byte, len(testData))
	_, err = rw.Read(buf)
	if err != nil {
		t.Errorf("Error reading through interface: %v", err)
	}
	t.Logf("Read through interface: %s", buf)

	// Обрабатываем данные через функцию, принимающую интерфейс
	result, err := Process(readWriter, []byte("process test"))
	if err != nil {
		t.Errorf("Error in Process: %v", err)
	}
	t.Logf("Process result: %s", result)
}

func TestFullInterface(t *testing.T) {
	// Создаем полную реализацию всех интерфейсов
	full := &FullImplementation{
		CustomReadWriter: CustomReadWriter{
			CustomReader: CustomReader{
				data: []byte("full implementation test"),
			},
		},
	}

	// Проверяем как ReadWriteCloser
	var rwc ReadWriteCloser = full

	// Записываем данные
	_, err := rwc.Write([]byte("more data"))
	if err != nil {
		t.Errorf("Error writing to full implementation: %v", err)
	}

	// Читаем данные
	buf := make([]byte, 4)
	_, err = rwc.Read(buf)
	if err != nil {
		t.Errorf("Error reading from full implementation: %v", err)
	}
	t.Logf("Read from full: %s", buf)

	// Закрываем
	err = rwc.Close()
	if err != nil {
		t.Errorf("Error closing: %v", err)
	}

	// Обрабатываем и закрываем через функцию
	result, err := ProcessAndClose(full, []byte("process and close"))
	if err != nil {
		t.Errorf("Error in ProcessAndClose: %v", err)
	}
	t.Logf("ProcessAndClose result: %s", result)

	// Проверяем различные типы приведения интерфейсов
	var reader Reader = full
	var writer Writer = full
	var closer Closer = full

	t.Logf("Interface conversions work: Reader: %T, Writer: %T, Closer: %T",
		reader, writer, closer)
}
