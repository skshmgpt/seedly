package prisma

import (
	"fmt"
	"regexp"
	"strings"
)

var modelStartPattern = regexp.MustCompile(`^model\s+([A-Za-z_][A-Za-z0-9_]*)\s*\{\s*$`)

func ParseString(input string) (Schema, error) {
	lines := strings.Split(input, "\n")
	var schema Schema

	for i := 0; i < len(lines); i++ {
		line := stripComment(lines[i])
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		match := modelStartPattern.FindStringSubmatch(trimmed)
		if match == nil {
			continue
		}

		model := Model{Name: match[1]}
		closed := false
		for i++; i < len(lines); i++ {
			fieldLine := strings.TrimSpace(stripComment(lines[i]))
			if fieldLine == "" {
				continue
			}
			if fieldLine == "}" {
				closed = true
				break
			}
			field, ok, err := parseField(fieldLine)
			if err != nil {
				return Schema{}, fmt.Errorf("parse %s.%s: %w", model.Name, fieldLine, err)
			}
			if ok {
				model.Fields = append(model.Fields, field)
			}
		}

		if !closed {
			return Schema{}, fmt.Errorf("model %s is missing closing brace", model.Name)
		}
		schema.Models = append(schema.Models, model)
	}

	if len(schema.Models) == 0 && strings.Contains(input, "model ") {
		return Schema{}, fmt.Errorf("no complete model blocks found")
	}

	return schema, nil
}

func parseField(line string) (Field, bool, error) {
	if strings.HasPrefix(line, "@@") {
		return Field{}, false, nil
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return Field{}, false, fmt.Errorf("field must include name and type")
	}

	typeName := parts[1]
	field := Field{Name: parts[0]}
	if strings.HasSuffix(typeName, "?") {
		field.IsOptional = true
		typeName = strings.TrimSuffix(typeName, "?")
	}
	if strings.HasSuffix(typeName, "[]") {
		field.IsList = true
		typeName = strings.TrimSuffix(typeName, "[]")
	}
	field.Type = typeName

	attrs := strings.Join(parts[2:], " ")
	field.IsID = strings.Contains(attrs, "@id")
	field.IsUnique = strings.Contains(attrs, "@unique")
	if value, ok := attrValue(attrs, "@default"); ok {
		field.HasDefault = true
		field.Default = value
	}
	if strings.Contains(attrs, "@relation") {
		relation, err := parseRelation(field.Type, attrs)
		if err != nil {
			return Field{}, false, err
		}
		field.Relation = &relation
	}

	return field, true, nil
}

func stripComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func attrValue(attrs, name string) (string, bool) {
	idx := strings.Index(attrs, name+"(")
	if idx < 0 {
		return "", false
	}
	start := idx + len(name) + 1
	depth := 1
	for i := start; i < len(attrs); i++ {
		switch attrs[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return attrs[start:i], true
			}
		}
	}
	return "", false
}

func parseRelation(target string, attrs string) (Relation, error) {
	value, ok := attrValue(attrs, "@relation")
	if !ok {
		return Relation{}, fmt.Errorf("relation missing arguments")
	}
	fields, ok := listArg(value, "fields")
	if !ok || len(fields) == 0 {
		return Relation{}, fmt.Errorf("relation missing fields")
	}
	references, ok := listArg(value, "references")
	if !ok || len(references) == 0 {
		return Relation{}, fmt.Errorf("relation missing references")
	}
	return Relation{TargetModel: target, Fields: fields, References: references}, nil
}

func listArg(input, name string) ([]string, bool) {
	pattern := regexp.MustCompile(name + `\s*:\s*\[([^\]]+)\]`)
	match := pattern.FindStringSubmatch(input)
	if match == nil {
		return nil, false
	}
	items := strings.Split(match[1], ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values, true
}
