package postcards2diane

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	defaultFont string
	fonts       map[string]*truetype.Font
)

func init() {
	fonts = make(map[string]*truetype.Font)
	err := filepath.Walk("fonts", func(path string, info os.FileInfo, err error) error {
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
			return err
		}
		name := filepath.Base(path)
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
