package cmd

import (
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/viant/bintly/codegen"
)

//RunClient validates CLI options and triggers code generator
func RunClient(Version string, args []string) int {
	options := &codegen.Options{}
	_, err := flags.ParseArgs(options, args)
	if err != nil {
		log.Fatal(err)
		return 1
	}
	err = options.Validate()
	if err != nil {
		log.Printf("validation eror: %v", err)
		return 1
	}
	err = codegen.Generate(options)
	if err != nil {
		log.Printf("code generation error %v", err)
		return 1
	}
	return 0
}
