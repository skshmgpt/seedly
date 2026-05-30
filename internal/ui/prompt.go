package ui

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/skshmgpt/seedly/internal/prisma"
)

func SelectModels(in io.Reader, out io.Writer, models []prisma.Model) ([]string, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models found in schema")
	}
	_, _ = fmt.Fprintln(out, "Select models to seed by number, separated by commas:")
	for idx, model := range models {
		_, _ = fmt.Fprintf(out, "%d) %s\n", idx+1, model.Name)
	}
	_, _ = fmt.Fprint(out, "> ")

	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("no models selected")
	}

	parts := strings.Split(line, ",")
	selected := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		idx, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || idx < 1 || idx > len(models) {
			return nil, fmt.Errorf("invalid model selection %q", strings.TrimSpace(part))
		}
		name := models[idx-1].Name
		if !seen[name] {
			selected = append(selected, name)
			seen[name] = true
		}
	}
	return selected, nil
}
