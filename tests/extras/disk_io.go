package extras

import (
	"os"
	"path/filepath"
)

func DiskBurst() error {
	p := filepath.Join(os.TempDir(), "nukastress_disk_io.tmp")
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer os.Remove(p)
	defer f.Close()

	buf := make([]byte, 1<<20)
	for i := 0; i < 64; i++ {
		if _, err := f.Write(buf); err != nil {
			return err
		}
	}
	return f.Sync()
}
