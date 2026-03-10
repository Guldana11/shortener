package pool

import (
	"bytes"
	"testing"
)

// bytes.Buffer реализует Reset(), подходит как Resetter
type testBuffer struct {
	bytes.Buffer
}

func (b *testBuffer) Reset() {
	b.Buffer.Reset()
}

func TestNew(t *testing.T) {
	p := New(func() *testBuffer {
		return &testBuffer{}
	})
	if p == nil {
		t.Fatal("New вернул nil")
	}
}

func TestGetReturnsObject(t *testing.T) {
	p := New(func() *testBuffer {
		return &testBuffer{}
	})

	buf := p.Get()
	if buf == nil {
		t.Fatal("Get вернул nil")
	}
}

func TestPutResetsAndReuses(t *testing.T) {
	p := New(func() *testBuffer {
		return &testBuffer{}
	})

	buf := p.Get()
	buf.WriteString("hello")
	if buf.Len() == 0 {
		t.Fatal("буфер должен содержать данные")
	}

	p.Put(buf)

	// после Put объект сброшен (Reset вызван)
	// при повторном Get можем получить тот же объект (или новый)
	buf2 := p.Get()
	if buf2.Len() != 0 {
		t.Errorf("ожидали пустой буфер после Put, длина=%d", buf2.Len())
	}
}

func TestGetPutCycle(t *testing.T) {
	p := New(func() *testBuffer {
		return &testBuffer{}
	})

	for i := 0; i < 100; i++ {
		buf := p.Get()
		buf.WriteString("data")
		p.Put(buf)
	}
}
