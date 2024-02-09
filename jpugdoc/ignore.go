package jpugdoc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type IgnoreList map[string]bool

func registerIgnore(ignoreName string, ignores []string) {
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
		_, err := fmt.Fprintf(f, "%s\n", ig)
		if err != nil {
			return fmt.Errorf("failed to write ignore: %w", err)
		}
	}
	return nil
}

func loadIgnore(fileName string) IgnoreList {
	f, err := os.Open(fileName)
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
