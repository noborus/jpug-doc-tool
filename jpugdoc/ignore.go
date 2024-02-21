package jpugdoc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type IgnoreList map[string]bool

// loadIgnore loads a list of strings to ignore from the specified file.
func loadIgnore(fileName string) IgnoreList {
	ignoreName := filepath.Join(DicDir, fileName+".ignore")
	f, err := os.Open(ignoreName)
	if err != nil {
		return nil
	}
	defer f.Close()

	return readIgnore(f)
}

func readIgnore(f io.Reader) IgnoreList {
	ignores := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ignores[scanner.Text()] = true
	}
	return ignores
}

func registerIgnore(fileName string, ignores []string) {
	ignoreName := filepath.Join(DicDir, fileName+".ignore")
	f, err := os.OpenFile(ignoreName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := writeIgnore(f, ignores); err != nil {
		log.Fatal(err)
	}
}

func writeIgnore(f io.Writer, ignores []string) error {
	for _, ig := range ignores {
		_, err := fmt.Fprintf(f, "%s\n", stripNL(ig))
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
