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
		switch {
		case strings.HasPrefix(m, "http://") || strings.HasPrefix(m, "https://"):
			// URL import
		case strings.HasPrefix(m, "github.com/"):
			// GitHub Contents API import
		case strings.HasPrefix(m, "./"):
			// Local path escape
		case yamlSuffixPattern.MatchString(m):
			// File path with .yaml/.yml suffix
		case strings.Contains(m, "/"):
			return nil, nil, fmt.Errorf("invalid migration argument %q: migration name must not contain /", arg)
		default:
			// Migration name
		}
		migrations = append(migrations, m)
	}
	return migrations, yamlFiles, nil
}
