package openapigen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/ettle/strcase"
	"github.com/getkin/kin-openapi/openapi3"
)

type Field struct {
	Name     string
	Type     string
	JsonName string
	Required bool
	Fields   []Field
}

func GenerateGetterHeaders(fnName string, structName string, fields []Field) string {
	code := fmt.Sprintf(`
		func %s(c *gin.Context) (%s, error) {
			data := %s{}
			stringValue := ""
	`, fnName, structName, structName)

	for _, field := range fields {
		code += "\n"
		code += fmt.Sprintf(`  stringValue = c.GetHeader("%s")`, field.JsonName)

		if field.Required {
			code += fmt.Sprintf(`
				if stringValue == "" {
					return data, fmt.Errorf("header no encontrado: \"%s\"")
				}
			`, field.JsonName)
		}

		fmt.Println(field.Name, field.Type, field.Required)

		pointerPrefix := ""
		if !field.Required {
			pointerPrefix = "&"
		}

		switch field.Type {
		case "bool":
			{
				code += fmt.Sprintf(`
					lowerStringValue := strings.ToLower(stringValue)
					if lowerStringValue != "true" && lowerStringValue != "false" {
						return data, fmt.Errorf("error al obtener header booleano \"%s\"")
					}

					%s := lowerStringValue == "true"
					data.%s = %s%s
				`, field.JsonName, field.Name, field.Name, pointerPrefix, field.Name)
				break
			}
		case "int":
			{
				code += fmt.Sprintf(`
					%s, err := strconv.Atoi(stringValue)
					if err != nil {
						return data, fmt.Errorf("error al obtener header numérico \"%s\": %%v", err)
					}

					data.%s = %s%s
				`, field.Name, field.JsonName, field.Name, pointerPrefix, field.Name)
				break
			}
		default:
			{
				code += fmt.Sprintf("\n  data.%s = %sstringValue", field.Name, pointerPrefix)
				break
			}
		}

		code += "\n"
	}

	code += "\n  return data, nil"
	code += "\n}\n"
	return code
}

func GenerateGetterQuery(fnName string, structName string, fields []Field) string {
	code := fmt.Sprintf(`
		func %s(c *gin.Context) (%s, error) {
			data := %s{}
			stringValue := ""
	`, fnName, structName, structName)

	for _, field := range fields {
		code += "\n"
		code += fmt.Sprintf(`  stringValue = c.Query("%s")`, field.JsonName)

		if field.Required {
			code += fmt.Sprintf(`
				if stringValue == "" {
					return data, fmt.Errorf("query param no encontrado: \"%s\"")
				}
			`, field.JsonName)
		}

		fmt.Println(field.Name, field.Type, field.Required)

		pointerPrefix := ""
		if !field.Required {
			pointerPrefix = "&"
		}

		switch field.Type {
		case "bool":
			{
				code += fmt.Sprintf(`
					lowerStringValue := strings.ToLower(stringValue)
					if lowerStringValue != "true" && lowerStringValue != "false" {
						return data, fmt.Errorf("error al obtener query param booleano \"%s\"")
					}

					%s := lowerStringValue == "true"
					data.%s = %s%s
				`, field.JsonName, field.Name, field.Name, pointerPrefix, field.Name)
				break
			}
		case "int":
			{
				code += fmt.Sprintf(`
					%s, err := strconv.Atoi(stringValue)
					if err != nil {
						return data, fmt.Errorf("error al obtener query param numérico \"%s\": %%v", err)
					}

					data.%s = %s%s
				`, field.Name, field.JsonName, field.Name, pointerPrefix, field.Name)
				break
			}
		default:
			{
				code += fmt.Sprintf("\n  data.%s = %sstringValue", field.Name, pointerPrefix)
				break
			}
		}

		code += "\n"
	}

	code += "\n  return data, nil"
	code += "\n}\n"
	return code
}

func GenerateGetterBody(fnName string, structName string) string {
	code := fmt.Sprintf(`
		func %s(c *gin.Context) (%s, error) {
			data := %s{}
			err := parseBody(c, &data)
			return data, err
		}
	`, fnName, structName, structName)

	return code
}

func FieldToString(f Field) string {
	fieldType := ""
	if !f.Required {
		fieldType += "*"
	}

	if len(f.Fields) == 0 {
		fieldType += translateTypeToGo(f.Type)
	} else {
		structType := "struct {\n"

		sort.SliceStable(f.Fields, func(i, j int) bool {
			return f.Fields[i].Name < f.Fields[j].Name
		})

		for i, subfield := range f.Fields {
			if i > 0 {
				structType += "\n"
			}

			structType += "  " + FieldToString(subfield)
		}

		structType += "\n}"

		fieldType = structType
	}

	s := fmt.Sprintf(
		"%s %s `json:\"%s\"`",
		f.Name,
		fieldType,
		f.JsonName,
	)

	return s
}

func translateTypeToGo(t string) string {
	switch t {
	case "integer":
		return "int"
	case "float":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return t
	}
}

func UpperCaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}

	head := string(s[0])
	tail := s[1:]
	head = strings.ToUpper(head)
	return head + tail
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ParseOperation(prefix string, operation *openapi3.Operation) (string, error) {
	headerFields := []Field{}
	queryFields := []Field{}

	for _, param := range operation.Parameters {
		field := parseParam(param)

		if param.Value.In == "header" {
			headerFields = append(headerFields, field)
			continue
		}

		if param.Value.In == "query" {
			queryFields = append(queryFields, field)
			continue
		}
	}

	bodyFields := []Field{}
	if operation.RequestBody != nil {
		jsonContent := operation.RequestBody.Value.Content.Get("application/json")

		if jsonContent != nil {
			fields := parseSchema(jsonContent.Schema)
			bodyFields = append(bodyFields, fields...)
		}
	}

	responseFields := []Field{}
	responseType := ""
	for responseIndex, response := range operation.Responses.Map() {
		jsonContent := response.Value.Content.Get("application/json")

		if jsonContent == nil {
			continue
		}

		currentResponseType := getParamType(jsonContent.Schema)

		if responseType == "" {
			responseType = currentResponseType
		}

		if responseType != currentResponseType {
			return "", fmt.Errorf("multiple response types not allowed. Found %q and %q", responseType, currentResponseType)
		}

		fields := parseSchema(jsonContent.Schema)

		for _, newField := range fields {
			previousFieldIndex := slices.IndexFunc(responseFields, func(f Field) bool {
				return f.JsonName == newField.JsonName
			})

			if previousFieldIndex < 0 {
				responseFields = append(responseFields, newField)
				continue
			}

			if responseFields[previousFieldIndex].Type != newField.Type {
				return "", fmt.Errorf(
					"%s, response %q, property %q has multiple types: found %q and %q",
					prefix,
					responseIndex,
					newField.Name,
					responseFields[previousFieldIndex].Type,
					newField.Type,
				)
			}

			if !newField.Required {
				responseFields[previousFieldIndex].Required = newField.Required
			}
		}
	}

	pkg := "package server\n"

	imports := `
		import (
			"fmt"
			"strconv"
			"strings"

			"github.com/gin-gonic/gin"
		)
	`

	var buf bytes.Buffer
	buf.WriteString(pkg)
	buf.WriteString(imports)

	if len(headerFields) > 0 {
		nameStructHeaders := prefix + "_Headers"
		codeHeaders := GenerateStructCode(nameStructHeaders, headerFields)
		buf.WriteString(codeHeaders)

		nameGetterHeaders := "GetHeaders_" + prefix
		codeGetterHeaders := GenerateGetterHeaders(nameGetterHeaders, nameStructHeaders, headerFields)
		buf.WriteString(codeGetterHeaders)
	}

	if len(queryFields) > 0 {
		nameStructQuery := prefix + "_Query"
		codeQuery := GenerateStructCode(nameStructQuery, queryFields)
		buf.WriteString(codeQuery)

		nameGetterQuery := "GetQuery_" + prefix
		codeGetterQuery := GenerateGetterQuery(nameGetterQuery, nameStructQuery, queryFields)
		buf.WriteString(codeGetterQuery)
	}

	if len(bodyFields) > 0 {
		nameStructBody := prefix + "_Body"
		codeBody := GenerateStructCode(nameStructBody, bodyFields)
		buf.WriteString(codeBody)

		nameGetterBody := "GetBody_" + prefix
		codeGetterBody := GenerateGetterBody(nameGetterBody, nameStructBody)
		buf.WriteString(codeGetterBody)
	}

	if len(responseFields) > 0 {
		nameStructResponse := prefix + "_Response"
		codeResponse := GenerateStructCode(nameStructResponse, responseFields)
		buf.WriteString(codeResponse)
	}

	bs := buf.Bytes()
	src := string(bs)
	bs, err := format.Source(bs)
	if err != nil {
		return src, fmt.Errorf("failed to format source code: %v", err)
	}

	src = string(bs)
	return src, nil
}

func parseParam(param *openapi3.ParameterRef) Field {
	field := Field{
		Name:     UpperCaseFirstLetter(strcase.ToCamel(param.Value.Name)),
		JsonName: param.Value.Name,
		Type:     "",
		Required: param.Value.Required,
		Fields:   []Field{},
	}

	types := *param.Value.Schema.Value.Type
	if len(types) > 0 {
		field.Type = translateTypeToGo(types[0])
	}

	return field
}

func GenerateStructCode(name string, fields []Field) string {
	if len(fields) == 1 && fields[0].Name == "" {
		code := fmt.Sprintf(`
			type %s %s
		`, name, fields[0].Type)
		return code
	}

	code := fmt.Sprintf("\ntype %s struct {", name)

	for _, field := range fields {
		code += "\n  " + FieldToString(field)
	}

	code += "\n}\n"
	return code
}

func getParamType(param *openapi3.SchemaRef) string {
	if param.Value.Type == nil {
		return ""
	}

	paramType := ""
	types := *param.Value.Type
	if len(types) > 0 {
		paramType += types[0]
	}

	return paramType
}

func parseField(paramOriginalName string, param *openapi3.SchemaRef, requiredParams []string) Field {
	field := Field{}

	field.Name = UpperCaseFirstLetter(strcase.ToCamel(paramOriginalName))
	field.JsonName = paramOriginalName
	field.Type = getParamType(param)

	if field.Type == "array" {
		types := param.Value.Items.Value.Type.Slice()
		arrayType := ""
		if len(types) > 0 {
			arrayType = types[0]
		}

		arrayType = translateTypeToGo(arrayType)

		if arrayType == "object" {
			arrayType = parseArrayType(param)
		}

		field.Type = "[]" + translateTypeToGo(arrayType)
	}

	for _, requiredParam := range requiredParams {
		if requiredParam == paramOriginalName {
			field.Required = true
			break
		}
	}

	for propName, prop := range param.Value.Properties {
		subfield := parseField(propName, prop, param.Value.Required)
		field.Fields = append(field.Fields, subfield)
	}

	return field
}

func parseArrayType(param *openapi3.SchemaRef) string {
	s := "struct {\n"

	props := param.Value.Items.Value.Properties
	required := param.Value.Items.Value.Required
	for propName, prop := range props {
		field := parseField(propName, prop, required)
		s += "\n  " + FieldToString(field)
	}

	s += "\n}"

	return s
}

func parseSchema(schema *openapi3.SchemaRef) []Field {
	fields := []Field{}

	paramType := ""
	if len(*schema.Value.Type) != 0 {
		paramType = (*schema.Value.Type)[0]
	}

	switch paramType {
	case "object":
		{
			keys := []string{}
			for key := range schema.Value.Properties {
				keys = append(keys, key)
			}

			sort.SliceStable(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})

			for _, key := range keys {
				param := schema.Value.Properties[key]
				field := parseField(key, param, schema.Value.Required)
				fields = append(fields, field)
			}

			break
		}
	case "array":
		{
			field := parseField("", schema, []string{})
			field.Required = true
			fields = append(fields, field)
			break
		}
	}

	return fields
}
