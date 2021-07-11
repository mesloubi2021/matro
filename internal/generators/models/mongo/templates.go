package mongo

const storeTpl = `
package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DB{{title (plural .Name) -}} = "{{- plural .Name }}"
	{{- range .Types }}
	Collection{{title (plural .Name)}} = "{{- plural .Name }}"
	{{- end}}
)

type {{ title .Name -}}Store struct {
	client *mongo.Client
}

// New {{- title .Name -}}Store makes connection to mongo server by provided url 
// and return an instance of the client
func New {{- title .Name -}}Store(ctx context.Context, url string) (* {{ title .Name -}}Store, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err 
	}
	return &{{- title .Name -}}Store {
		client: client,
	}, nil
} 

func (s *{{- title .Name -}}Store) getCollection (collection string) *mongo.Collection {
	return s.client.Database(DB{{ title (plural .Name) -}}).Collection(collection)
}
`

const crudTpl = `
package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"{{- .Repo -}}/internal/models/{{- .Module -}}store"


)

func (s *{{- title .Module -}}Store) Create{{- title .Name -}} (ctx context.Context, {{.Name}} *{{-  .Module -}}store. {{- title .Name}}) (* {{ .Module -}}store. {{- title .Name}}, error) {
	
	if _, err := s.getCollection(Collection{{title (plural .Name)}}).InsertOne(ctx, {{.Name}}); err != nil {
		return nil, err
	}
	return {{ .Name}}, nil
}

func (s *{{- title .Module -}}Store) Get{{- title .Name -}}ByID (ctx context.Context, id string) (* {{ .Module -}}store. {{- title .Name}}, error) {
	query := bson.M{
		"_id": id,
	}
	var {{.Name}} {{ .Module -}}store. {{- title .Name}}
	if err := s.getCollection(Collection{{title (plural .Name)}}).FindOne(ctx, query).Decode(&{{- .Name -}}); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &{{- .Name}}, nil
}

func (s *{{- title .Module -}}Store) Get{{- title (plural .Name) -}} (
	ctx context.Context,
	{{- if not (empty .SearchFields)}}
	search *string,
	{{- end}}
	{{- template "getargs" . }}
	offset *string,
	limit *int, 
) (
	[]*{{ .Module -}}store. {{- title .Name}}, 
	error,
) {
	query := bson.M{}
	{{- template "searchQuery" .}}

	var {{plural .Name}} []*{{ .Module -}}store. {{- title .Name}}
	if err := s.getCollection(Collection{{title (plural .Name)}}).Find(ctx, query).Decode(&{{- plural .Name -}}); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &{{- plural .Name}}, nil
}

func (s *{{- title .Module -}}Store) Update{{- title .Name -}} (ctx context.Context, {{.Name}} *{{-  .Module -}}store. {{- title .Name}}) (* {{ .Module -}}store. {{- title .Name}}, error) {
	query := bson.M{
		"_id": {{.Name -}}.Id,
	}
	if _, err := s.getCollection(Collection{{title (plural .Name)}}).UpdateOne(ctx, query, {{.Name}}); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return {{ .Name}}, nil
}
func (s *{{- title .Module -}}Store) Delete{{- title .Name -}}ByID (ctx context.Context, id string) (error) {
	query := bson.M{
		"_id": id,
	}
	if _,  err := s.getCollection(Collection{{title (plural .Name)}}).DeleteOne(ctx, query); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return  err
	}
	return nil
}

{{- define "getargs"}}
{{- $t := .}}
{{- range .Filters}}
	{{.}} *{{$t.FieldType .}},
{{- end}}
{{- end}}

{{- define "searchQuery" }}
{{- if not (empty .SearchFields)}}
	searchQuery := map[string][]interface{}{
		"$or": []interface{
			{{- range .SearchFields }}
			map[string]interface{}{"{{- . -}}": search},
			{{- end}}
		}
	}
{{- end}}
{{- end}}
`

const modelTyp = `
package {{ .Module -}}store

type {{title .Name}} struct {
	{{- counter 0}} 
	{{- range .Fields}}
	{{- if  (not (and .IsJoinedData  .IsList))}}
	{{ .GoName}}  {{- .GoType}}` + "`bson:\"{{- .Name}}\"`" + `  
	{{- end}}
	{{- end}}
}`
