package colourise

import (
	"fmt"
	"strings"

	"github.com/aws-cloudformation/rain/console/text"
)

func Yaml(in string) string {
	parts := strings.Split(in, "\n")

	for i, part := range parts {
		line := strings.Split(part, ":")

		if len(line) == 2 {
			splitLine := strings.Split(line[1], "  #")

			if len(splitLine) == 2 {
				parts[i] = fmt.Sprintf("%s:%s  %s", line[0], text.Yellow(splitLine[0]), text.Grey("#"+splitLine[1]))
			} else {
				parts[i] = fmt.Sprintf("%s:%s", line[0], text.Yellow(line[1]))
			}
		}
	}

	return strings.Join(parts, "\n")
}
