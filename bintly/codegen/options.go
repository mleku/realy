package codegen

import (
	"errors"
	"strings"
)

//Options represents Cli options
type Options struct {
	Source string   `short:"s" long:"sourceURL" description:"source URL"`
	Dest   string   `short:"d" long:"destinationURL" description:"destination URL"`
	Types  []string `short:"t" long:"struct type" description:"struct type"`
	Pkg    string   `short:"p" long:"package" description:"package"`
}

//Validate validates input options
func (o *Options) Validate() error {
	if o == nil {
		return errors.New("options is nil")
	}
	if o.Source == "" {
		return errors.New("source is empty")
	}
	if o.Types == nil || len(o.Types) == 0 {
		return errors.New("type is empty")
	}
	var srcPrefix = o.Dest
	if o.Dest == "" {
		srcPrefix = o.Source[:strings.LastIndex(o.Source, "/")]
	}
	o.Dest = srcPrefix + strings.Split(o.Source[strings.LastIndex(o.Source, "/"):], ".")[0] + "_enc.go"
	return nil
}
