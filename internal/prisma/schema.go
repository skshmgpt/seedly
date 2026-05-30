package prisma

type Schema struct {
	Models []Model
}

func (s Schema) Model(name string) (Model, bool) {
	for _, model := range s.Models {
		if model.Name == name {
			return model, true
		}
	}
	return Model{}, false
}

type Model struct {
	Name   string
	Fields []Field
}

func (m Model) Field(name string) (Field, bool) {
	for _, field := range m.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return Field{}, false
}

func (m Model) PrimaryKeyFields() []Field {
	var fields []Field
	for _, field := range m.Fields {
		if field.IsID {
			fields = append(fields, field)
		}
	}
	return fields
}

type Field struct {
	Name       string
	Type       string
	IsOptional bool
	IsList     bool
	IsID       bool
	IsUnique   bool
	HasDefault bool
	Default    string
	Relation   *Relation
}

func (f Field) IsScalar() bool {
	switch f.Type {
	case "String", "Int", "BigInt", "Float", "Decimal", "Boolean", "DateTime", "Json", "Bytes":
		return true
	default:
		return false
	}
}

type Relation struct {
	TargetModel string
	Fields      []string
	References  []string
}
