package postcards2diane

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	defaultFont string
	fonts       map[string]*truetype.Font
)

func init() {
	dir := os.Getenv("POSTCARD_FONTS_DIR")
	if dir == "" {
		dir = "/Library/fonts" // Mac OS X default
	}
	fonts = make(map[string]*truetype.Font)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".ttf" {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		ttf, err := freetype.ParseFont(b)
		if err != nil {
			return nil // skip, don't error tho
		}
		name := strings.TrimSuffix(filepath.Base(path), ".ttf")
		fonts[name] = ttf
		if defaultFont == "" {
			defaultFont = name
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
