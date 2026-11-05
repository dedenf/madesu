package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type entry struct {
	name  string
	isDir bool
}

// readDir mengembalikan isi dir: folder dulu, lalu file, dotfile disembunyikan.
// Entry ".." ditaruh di paling atas kalau bukan root.
func readDir(dir string) ([]entry, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]entry, 0, len(des))
	for _, de := range des {
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		out = append(out, entry{name: de.Name(), isDir: de.IsDir()})
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
