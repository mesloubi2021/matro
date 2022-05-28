package types

import (
	"fmt"
	"strings"

	"github.com/firstcontributions/matro/internal/generators/utils"
	"github.com/firstcontributions/matro/internal/parser"
	"github.com/gertd/go-pluralize"
)

// Field defines the field meta data by its type, is it a list,
// is it nullable etc..
type Field struct {
	Name             string
	Type             string
	IsList           bool
	IsNullable       bool
	IsPaginated      bool
	IsQuery          bool
	Args             []Field
	IsPrimitive      bool
	IsJoinedData     bool
	IsMutatable      bool
	HardcodedFilters map[string]string
	NoGraphql        bool
}

// paginationArgs are the defualt pagination arguments should be
//  there with graphql relay paginated queries
var paginationArgs = []Field{
	{Name: "first", Type: parser.Int},
	{Name: "last", Type: parser.Int},
	{Name: "after", Type: parser.String},
	{Name: "before", Type: parser.String},
}

// TODO(@gokultp) clean up this function, make it more readable
// NewField returns an instance of the field
func NewField(typesMap map[string]*parser.Type, typeDef *parser.Type, name string) *Field {
	if typeDef.IsPrimitive() {
		return &Field{
			Name:        name,
			Type:        typeDef.Type,
			IsPrimitive: true,
		}
	}

	f := &Field{
		Name:             name,
		IsList:           typeDef.Type == parser.List,
		IsPaginated:      typeDef.Paginated,
		IsJoinedData:     typeDef.JoinedData,
		HardcodedFilters: typeDef.HardcodedFilters,
	}
	if typeDef.Schema == "" {
		f.Type = typeDef.Name
		f.NoGraphql = typeDef.NoGraphql
	} else if !IsCompositeType(typeDef.Schema) {
		f.Type = typeDef.Schema
		f.NoGraphql = typeDef.NoGraphql
	} else {
		f.Type = typesMap[typeDef.Schema].Name
		f.NoGraphql = typesMap[typeDef.Schema].NoGraphql
	}
	if _, ok := typesMap[f.Type]; ok {
		f.Args = getArgs(typesMap, typesMap[f.Type])
	}

	if f.IsPaginated {
		f.Args = append(f.Args, paginationArgs...)
		f.IsQuery = true
	}
	return f
}

// getArgs gets argumets for query
func getArgs(typesMap map[string]*parser.Type, typeDef *parser.Type) []Field {
	args := []Field{}
	for _, a := range typeDef.Meta.Filters {

		for pName, pType := range typesMap[typeDef.Name].Properties {
			if pName == a {
				if pType.IsPrimitive() {
					args = append(args, Field{
						Name:        a,
						Type:        pType.Type,
						IsPrimitive: true,
					})
				} else {
					args = append(args, Field{
						Name:        a,
						Type:        parser.String,
						IsPrimitive: false,
					})
				}
				break
			}
		}
	}
	return args
}

// GoName return the field name to be used in go code
func (f *Field) GoName(allExported ...bool) string {
	exported := len(allExported) > 0 && allExported[0]
	if f.Name == "id" {
		return "Id"
	}
	if !exported && f.IsJoinedData {
		return utils.ToCamelCase(f.Name)
	}
	return utils.ToTitleCase(f.Name)
}

// GoType return the gotype to be used in go code
// args[0] graphql enabled
// args[1] update type
func (f *Field) GoType(args ...bool) string {
	var t string
	if f.IsJoinedData {
		t = "string"
	} else {
		t = GetGoType(f.Type)
		if len(args) > 0 && args[0] {
			t = GetGoGraphQLType(f.Type)
		}
		if f.IsList {
			t = "[]" + t
		}
	}
	if (f.IsPrimitive || f.Type == "time") && len(args) > 1 && args[1] {
		t = "*" + t
	}

	return t
}

// GraphQLFormattedName returns the formatted graphql name for the field
// if it is queiriable it formats like field(args...):Type!
func (f *Field) GraphQLFormattedName() string {
	name := utils.ToCamelCase(f.Name)
	if !f.IsQuery || len(f.Args) == 0 {
		return name
	}
	args := []string{}
	for _, a := range f.Args {
		if _, ok := f.HardcodedFilters[a.Name]; ok {
			// no need to add hardcoded filters in graphql query args
			continue
		}
		args = append(args, fmt.Sprintf("%s: %s", utils.ToCamelCase(a.Name), GetGraphQLType(&a)))
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(args, ", "))
}

// GraphQLFortmattedType return the graphql type name
func (f *Field) GraphQLFortmattedType() string {
	t := GetGraphQLType(f)
	if f.IsPaginated {
		plType := pluralize.NewClient().Plural(f.Type)
		t = fmt.Sprintf("%sConnection", utils.ToTitleCase(plType))
	}
	if f.IsList && !f.IsPaginated {
		t = fmt.Sprintf("[%s]", t)
	}
	if !f.IsNullable {
		t = t + "!"
	}
	return t
}

func (f *Field) ArgNames() []string {
	args := []string{}
	for _, a := range f.Args {
		args = append(args, a.Name)
	}
	return args
}
