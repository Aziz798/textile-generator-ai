package catalog

import (
	"path/filepath"
	"strconv"
	"strings"
)

func IsImageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

type NameRegistry struct {
	used map[string]int
}

func NewNameRegistry() *NameRegistry {
	return &NameRegistry{used: make(map[string]int)}
}

func (r *NameRegistry) OutputName(name string) string {
	r.used[name]++
	if r.used[name] == 1 {
		return name
	}
	return name + "-" + strconv.Itoa(r.used[name])
}
