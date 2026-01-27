package audit

import (
	"encoding/json"
	"os"
)

type FileObserver struct {
	path string
}

func NewFileObserver(path string) *FileObserver {
	return &FileObserver{path: path}
}

func (f *FileObserver) Notify(e Event) {
	file, err := os.OpenFile(f.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	data, _ := json.Marshal(e)
	file.Write(append(data, '\n'))
}
