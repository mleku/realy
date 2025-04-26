package codegen

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/viant/toolbox"
)

//session stores generated codes, imports, pkg
type session struct {
	*Options
	*toolbox.FileSetInfo
	pkg                 string
	structCodingCode    []string
	generatedTypes      map[string]bool
	imports             map[string]bool
	isBlockTemplateDone bool
}

//shallGenerateCode stores generated codes for all types
func (s *session) shallGenerateCode(typeName string) bool {
	if _, ok := s.generatedTypes[typeName]; ok {
		return false
	}
	s.generatedTypes[typeName] = true
	return true
}

//readPackageCode creates pkg code
func (s *session) readPackageCode() error {
	p, err := filepath.Abs(s.Source)
	if err != nil {
		return err
	}

	var f os.FileInfo
	if f, err = os.Stat(p); err != nil {
		// path/to/whatever does not exist
		return err
	}

	if !f.IsDir() {
		dir, _ := filepath.Split(p)
		_, pkg := path.Split(filepath.Base(dir))
		s.Pkg = pkg
		s.FileSetInfo, err = toolbox.NewFileSetInfo(dir)

	} else {
		_, pkg := path.Split(filepath.Base(p))
		s.Pkg = pkg
		s.FileSetInfo, err = toolbox.NewFileSetInfo(p)
	}

	// if Pkg flag is set use it
	if s.Pkg != "" {
		s.pkg = s.Pkg
	}

	return err
}

//addImports adds imports
func (s *session) addImport(pkg string) {
	s.imports[`"`+pkg+`"`] = true
}

//getImports returns imports
func (s *session) getImports() string {
	return "\t" + strings.Join(toolbox.MapKeysToStringSlice(s.imports), "\n\t")
}

//newSession creates a new session
func newSession(option *Options) *session {
	return &session{Options: option,
		structCodingCode: make([]string, 0),
		generatedTypes:   make(map[string]bool),
		imports:          make(map[string]bool),
	}
}
