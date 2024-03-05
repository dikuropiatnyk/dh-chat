package communication

import (
	"bufio"
	"fmt"
	"strings"
)

func GetInput(prompt string, reader *bufio.Reader) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(input, "\n"), nil
}
