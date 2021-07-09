package proto

import (
	"fmt"
	"os/exec"

	"github.com/firstcontributions/matro/internal/generators/types"
	"github.com/firstcontributions/matro/internal/parser"
	"github.com/firstcontributions/matro/internal/writer"
)

// Generator implements gRPC protobuf generator
type Generator struct {
	*types.TypeDefs
	modules map[string]Module
	Path    string
}

// Module encapsulates module meta data and types in module
type Module struct {
	parser.Module
	Types []*types.CompositeType
}

// NewGenerator returns an instance of gRPC code generator
func NewGenerator(path string, d *parser.Definition) *Generator {
	td := types.NewTypeDefs(path, d)
	mods := map[string]Module{}
	for _, m := range d.Modules {
		if m.DataSource == "grpc" {
			mods[m.Name] = Module{
				Module: m,
				Types:  td.GetTypeDefs(m.Entities),
			}
		}
	}
	return &Generator{
		TypeDefs: td,
		modules:  mods,
		Path:     path,
	}
}

// Generate generates gRPC prtobuf code for all given modules(services)
func (g *Generator) Generate() error {
	for _, m := range g.modules {
		if err := g.generateProtoForModule(m, tmpl); err != nil {
			return err
		}
		if err := g.generateGRPCService(g.protoFilePathForModule(m)); err != nil {
			return err
		}
	}
	return nil
}

// protoFilePathForModule returns path to protofile to be generated
func (g *Generator) protoFilePathForModule(m Module) string {
	return fmt.Sprintf("%s/api/%s.proto", g.Path, m.Name)
}

// generateProtoForModule generates proto file for the given module
func (g *Generator) generateProtoForModule(m Module, tmpl string) error {
	path := fmt.Sprintf("%s/api", g.Path)
	return writer.CompileAndWrite(
		path,
		m.Name+".proto",
		tmpl,
		m,
	)
}

// generateGRPCService generates grpc service stub from the proto file
func (g *Generator) generateGRPCService(protoPath string) error {
	if _, err := exec.Command(
		"protoc",
		protoPath,
		"--go_out="+g.Path,
		"--go-grpc_out="+g.Path,
	).Output(); err != nil {
		return err
	}
	return nil
}
