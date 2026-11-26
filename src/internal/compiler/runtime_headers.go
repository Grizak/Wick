package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Grizak/Wick/src/internal/ast"
	"github.com/Grizak/Wick/src/internal/types"
)

var (
	cHeaderComments        = regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*`)
	cHeaderDirectives      = regexp.MustCompile(`(?m)^[ \t]*#.*(?:\n|$)`)
	cFunctionName          = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*$`)
	cParameterNameAndArray = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*(\[\s*\])?\s*$`)
)

func parseRuntimeHeaders(includeDir string) ([]ast.NodeExtern, error) {
	files, err := filepath.Glob(filepath.Join(includeDir, "*.h"))
	if err != nil {
		return nil, fmt.Errorf("find runtime headers in %q: %w", includeDir, err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no runtime headers found in %q", includeDir)
	}
	sort.Strings(files)

	var functions []ast.NodeExtern
	signatures := make(map[string]string)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read runtime header %q: %w", file, err)
		}
		declarations := cHeaderComments.ReplaceAllStringFunc(string(content), replaceWithSpaces)
		declarations = cHeaderDirectives.ReplaceAllStringFunc(declarations, replaceWithSpaces)
		offset := 0
		for _, rawDeclaration := range strings.Split(declarations, ";") {
			declaration := strings.TrimSpace(rawDeclaration)
			if declaration != "" {
				declarationOffset := offset + strings.Index(rawDeclaration, declaration)
				function, err := parseCFunctionDeclaration(file, declarations, declarationOffset, declaration)
				if err != nil {
					return nil, err
				}
				if function != nil {
					signature := externSignature(function)
					if previous, exists := signatures[function.Name]; exists {
						if previous != signature {
							return nil, fmt.Errorf("%s:%d: conflicting declarations for runtime function %q",
								file, function.Pos.Line, function.Name)
						}
					} else {
						signatures[function.Name] = signature
						functions = append(functions, *function)
					}
				}
			}
			offset += len(rawDeclaration) + 1
		}
	}
	return functions, nil
}

func parseCFunctionDeclaration(file, source string, offset int, declaration string) (*ast.NodeExtern, error) {
	openParen := strings.IndexByte(declaration, '(')
	if openParen < 0 {
		return nil, nil
	}
	if !strings.HasSuffix(strings.TrimSpace(declaration), ")") {
		return nil, headerError(file, source, offset, "unsupported function declaration")
	}

	prefix := strings.TrimSpace(declaration[:openParen])
	nameMatch := cFunctionName.FindStringIndex(prefix)
	if nameMatch == nil {
		return nil, headerError(file, source, offset, "cannot determine function name")
	}
	name := prefix[nameMatch[0]:nameMatch[1]]
	returnType, err := cTypeToWick(prefix[:nameMatch[0]], 0)
	if err != nil {
		return nil, headerError(file, source, offset, fmt.Sprintf("function %q return type: %v", name, err))
	}

	paramsText := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(declaration), ")")[openParen+1:])
	if strings.ContainsAny(paramsText, "()") {
		return nil, headerError(file, source, offset, fmt.Sprintf("function %q uses an unsupported nested declarator", name))
	}
	var params []ast.NodeParam
	if paramsText != "" && paramsText != "void" {
		for i, rawParam := range strings.Split(paramsText, ",") {
			param, err := parseCParameter(rawParam, i)
			if err != nil {
				return nil, headerError(file, source, offset, fmt.Sprintf("function %q parameter: %v", name, err))
			}
			params = append(params, param)
		}
	}

	position := headerPosition(file, source, offset)
	return &ast.NodeExtern{
		Name:       name,
		Params:     &params,
		ReturnType: &returnType,
		Pos:        position,
	}, nil
}

func parseCParameter(declaration string, index int) (ast.NodeParam, error) {
	declaration = strings.TrimSpace(declaration)
	match := cParameterNameAndArray.FindStringSubmatchIndex(declaration)
	name := fmt.Sprintf("arg%d", index)
	typeDeclaration := declaration
	arrayDepth := 0
	if match != nil && strings.TrimSpace(declaration[:match[0]]) != "" {
		name = declaration[match[2]:match[3]]
		typeDeclaration = declaration[:match[0]]
		if match[4] >= 0 {
			arrayDepth = 1
		}
	}
	typeName, err := cTypeToWick(typeDeclaration, arrayDepth)
	if err != nil {
		return ast.NodeParam{}, err
	}
	return ast.NodeParam{Name: name, Type: &typeName}, nil
}

func cTypeToWick(cType string, arrayDepth int) (string, error) {
	fields := strings.Fields(strings.ReplaceAll(cType, "*", " * "))
	pointerDepth := arrayDepth
	var baseType []string
	for _, field := range fields {
		switch field {
		case "const", "volatile", "restrict", "__restrict", "extern", "static", "inline", "register":
			continue
		case "*":
			pointerDepth++
		default:
			baseType = append(baseType, field)
		}
	}

	var wickType string
	switch strings.Join(baseType, " ") {
	case "void":
		wickType = "void"
	case "char", "signed char", "unsigned char":
		wickType = "i8"
	case "int", "signed int", "unsigned int":
		wickType = "i32"
	case "long long", "signed long long", "unsigned long long":
		wickType = "i64"
	case "double":
		wickType = "f64"
	default:
		return "", fmt.Errorf("unsupported C type %q", strings.TrimSpace(cType))
	}
	if wickType == "void" && pointerDepth > 0 {
		wickType = "i8"
	}
	return strings.Repeat("*", pointerDepth) + wickType, nil
}

func replaceWithSpaces(text string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return r
		}
		return ' '
	}, text)
}

func externSignature(function *ast.NodeExtern) string {
	var signature strings.Builder
	signature.WriteString(*function.ReturnType)
	signature.WriteByte('(')
	if function.Params != nil {
		for i, param := range *function.Params {
			if i > 0 {
				signature.WriteByte(',')
			}
			signature.WriteString(*param.Type)
		}
	}
	signature.WriteByte(')')
	return signature.String()
}

func headerPosition(file, source string, offset int) types.Position {
	if offset > len(source) {
		offset = len(source)
	}
	prefix := source[:offset]
	lineStart := strings.LastIndexByte(prefix, '\n') + 1
	return types.Position{
		Filename: file,
		Line:     strings.Count(prefix, "\n") + 1,
		Column:   offset - lineStart + 1,
		Index:    offset,
	}
}

func headerError(file, source string, offset int, message string) error {
	position := headerPosition(file, source, offset)
	return fmt.Errorf("%s:%d:%d: %s", file, position.Line, position.Column, message)
}
