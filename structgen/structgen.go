package structgen

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

func getJsonTag(s string) string {
	cleanTag := strings.Trim(s, "`")

	tag := reflect.StructTag(cleanTag)
	return tag.Get("json")
}

type StructGenerator struct {
	BaseDir string
	PkgPath string
}

func (sg *StructGenerator) loadStructDefinition(structName string) *ast.StructType {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  sg.BaseDir,
	}
	pkgs, _ := packages.Load(cfg, sg.PkgPath)
	var dataStruct *ast.StructType
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				ts, ok := n.(*ast.TypeSpec)
				if !ok || ts.Name.Name != structName {
					return true
				}

				s, ok := ts.Type.(*ast.StructType)
				if !ok {
					return true
				}
				dataStruct = s
				return false
			})
		}
	}
	return dataStruct
}

func isBuiltinType(name string) bool {
	builtins := map[string]bool{
		"bool": true, "string": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"uintptr": true, "byte": true, "rune": true,
		"float32": true, "float64": true, "float16": true,
		"complex64": true, "complex128": true,
		"error": true,
	}
	return builtins[name]
}

func getFieldType(expr ast.Expr) (string, bool, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		isBuiltin := isBuiltinType(t.Name)
		return t.Name, false, isBuiltin
	case *ast.StarExpr:
		baseType, _, isStruct := getFieldType(t.X)
		return "*" + baseType, true, isStruct
	case *ast.ArrayType:
		baseType, _, isStruct := getFieldType(t.Elt)
		return "[]" + baseType, false, isStruct
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", t.X, t.Sel), false, false
	case *ast.MapType:
		keyType, _, _ := getFieldType(t.Key)
		valueType, _, _ := getFieldType(t.Value)
		return fmt.Sprintf("map[%s]%s", keyType, valueType), false, false
	default:
		return "unknown", false, false
	}
}

func sliceString(values []string) string {
	res := "{"
	for _, i := range values {
		res = res + fmt.Sprintf(`"%s",`, i)
	}
	res = res + "}"
	return res
}

func sliceNonString(typ string) bool {
	types := map[string]bool{
		"[]float64": true, "[]int64": true, "[]int32": true, "[]int16": true,
		"[]float32": true, "[]int": true, "[]uint64": true, "[]uint32": true,
		"[]uint16": true, "[]uint8": true, "[]int8": true,
	}
	return types[typ]
}

func (sg *StructGenerator) mapItem(field string, value any, ptr, builtin bool, r *ast.Field) string {
	var row string
	if strings.Contains(field, "time.time") {
		row = fmt.Sprintf("%s: %s,\n", r.Names[0], "time.Now()")
		return row
	}
	if field == "[]string" {
		if arr, ok := value.([]any); ok {
			strs := make([]string, len(arr))
			for i, v := range arr {
				strs[i] = fmt.Sprint(v)
			}
			row = fmt.Sprintf("%s: %s%s,\n", r.Names[0], field, sliceString(strs))
			return row
		}
	}

	if field == "map[string]any" {
		return row
	}

	if strings.HasPrefix(field, "[]") && !builtin {
		structType := strings.TrimPrefix(field, "[]")
		pkgPath := cleanPkgPath(sg.PkgPath)
		if arr, ok := value.([]any); ok {
			row = fmt.Sprintf("%s: []%s%s{\n", r.Names[0], pkgPath, structType)
			for _, item := range arr {
				if itemMap, ok := item.(map[string]any); ok {
					nestedStruct := sg.MapField(structType, itemMap)
					row += fmt.Sprintf("    %s,\n", nestedStruct)
				}
			}
			row += "},\n"
			return row
		}
	}

	if sliceNonString(field) {
		v := fmt.Sprintf("%v", value)
		v = strings.ReplaceAll(v, " ", ",")
		v = strings.ReplaceAll(v, "[", "")
		v = strings.ReplaceAll(v, "]", "")

		row = fmt.Sprintf("%s:%s{%s,},\n", r.Names[0], field, v)
		return row
	}

	if builtin {
		value = getCorrectValue(value)
		if !ptr {
			row = fmt.Sprintf("%s: %s,\n", r.Names[0], value)

			return row
		}
		if ptr {
			row = fmt.Sprintf("%s: %s,\n", r.Names[0], fmt.Sprintf("Ptr(%s)", value))

			return row
		}
	}

	if !builtin {
		var st string
		pkgPath := cleanPkgPath(sg.PkgPath)
		if dt, ok := value.(map[string]any); ok {
			st = sg.MapField(field, dt)
		}
		if !ptr {
			row = fmt.Sprintf("%s: %s%s%s,\n", r.Names[0], pkgPath, field, st)
			return row
		}
		if ptr {
			row = fmt.Sprintf("%s: Ptr(%s%s%s),\n", r.Names[0], pkgPath, field, st)

			return row
		}
	}
	return row
}

func cleanPkgPath(pkgPath string) string {
	if pkgPath == "./" {
		pkgPath = ""
	}
	pkgPath = strings.ReplaceAll(pkgPath, "./", "")
	pkgPath = pkgPath + "."
	return pkgPath
}

func getCorrectValue(value any) string {
	if data, ok := value.(int); ok {
		return fmt.Sprintf("%d", data)
	}
	if data, ok := value.(string); ok {
		return fmt.Sprintf(`"%s"`, data)
	}
	if data, ok := value.(bool); ok {
		return fmt.Sprintf(`%v`, data)
	}
	if data, ok := value.(float64); ok {
		return fmt.Sprintf("%v", data)
	}
	return ""
}

func (sg *StructGenerator) MapField(stName string, rawJson map[string]any) string {
	dataStruct := sg.loadStructDefinition(stName)
	var sb strings.Builder

	sb.WriteString("{\n")
	for _, r := range dataStruct.Fields.List {
		jsonTag := getJsonTag(r.Tag.Value)
		value, ok := rawJson[jsonTag]
		if !ok {
			continue
		}
		ft, ptr, builtin := getFieldType(r.Type)
		sb.WriteString(sg.mapItem(ft, value, ptr, builtin, r))

	}
	sb.WriteString("}")
	return sb.String()
}
