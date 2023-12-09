package jpugdoc

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type IgnoreList map[string]bool

func loadIgnore(fileName string) IgnoreList {
	f, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer f.Close()

	ignores := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ignores[scanner.Text()] = true
	}
	return ignores
}

func registerIgnore(fileName string, ignores []string) {
	ignoreName := DicDir + fileName + ".ignore"

	f, err := os.OpenFile(ignoreName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	for _, ig := range ignores {
		fmt.Fprintf(f, "%s\n", ig)
	}
}
