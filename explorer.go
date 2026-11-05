package main

import (
	"os"
	"path/filepath"
	"sort"
)

type entry struct {
	name   string
	isDir  bool // folder, termasuk symlink yang menunjuk ke folder
	isLink bool // symlink (ke file maupun folder, termasuk yang rusak)
}

// readDir mengembalikan semua isi dir: folder dulu, lalu file, diurutkan nama.
// Dotfile ikut ditampilkan. Symlink ke folder dianggap folder; symlink rusak tetap
// ditampilkan sebagai file. Entry ".." ditaruh di paling atas kalau bukan root.
func readDir(dir string) ([]entry, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]entry, 0, len(des)+1)
	for _, de := range des {
		e := entry{name: de.Name(), isDir: de.IsDir()}
		if de.Type()&os.ModeSymlink != 0 {
			e.isLink = true
			if fi, err := os.Stat(filepath.Join(dir, de.Name())); err == nil && fi.IsDir() {
				e.isDir = true
			}
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].isDir != out[j].isDir {
			return out[i].isDir
		}
		return out[i].name < out[j].name
	})
	if filepath.Dir(dir) != dir {
		out = append([]entry{{name: "..", isDir: true}}, out...)
	}
	return out, nil
}
