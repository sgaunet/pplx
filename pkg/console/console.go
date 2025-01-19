package console

import (
	"bufio"
	"fmt"
	"os"
)

func Input(label string) (string, error) {
	fmt.Printf("%s:\n", label)

	scanner := bufio.NewScanner(os.Stdin)
	var lines string
	for {
		scanner.Scan()
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		lines = lines + " " + line
	}

	err := scanner.Err()
	return lines, err
}
