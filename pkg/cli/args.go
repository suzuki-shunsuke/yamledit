package cli

import (
	"fmt"
	"regexp"
	"strings"
)

var yamlSuffixPattern = regexp.MustCompile(`\.ya?ml$`)

func parseArgs(args []string) (migrations []string, yamlFiles []string, err error) {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "@") {
			yamlFiles = append(yamlFiles, arg)
			continue
		}
		m := arg[1:]
		if yamlSuffixPattern.MatchString(m) {
			migrations = append(migrations, m)
			continue
		}
		if strings.Contains(m, "/") {
			return nil, nil, fmt.Errorf("invalid migration argument %q: migration name must not contain /", arg)
		}
		migrations = append(migrations, m)
	}
	return migrations, yamlFiles, nil
}
