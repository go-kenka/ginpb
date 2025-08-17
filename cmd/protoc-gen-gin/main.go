package main

import (
	"flag"
	"fmt"

	"github.com/go-kenka/ginpb/internal/gen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
	omitempty   = flag.Bool("omitempty", true, "omit if google.api is empty")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-gin %v\n", gen.Release)
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range plugin.Files {
			if !f.Generate {
				continue
			}

			gen.GenerateFile(plugin, f, *omitempty)
		}
		return nil
	})
}
