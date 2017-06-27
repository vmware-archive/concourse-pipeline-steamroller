package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"

	"github.com/concourse/atc"
	"github.com/jessevdk/go-flags"
	steamroller "github.com/krishicks/concourse-pipeline-steamroller"
	yamlpatch "github.com/krishicks/yaml-patch"
	yaml "gopkg.in/yaml.v2"
)

type opts struct {
	PipelinePath FileFlag `long:"pipeline" short:"p" value-name:"PATH" description:"Path to pipeline"`
	ConfigPath   FileFlag `long:"config" short:"c" value-name:"PATH" description:"Path to config"`
}

func main() {
	var o opts
	_, err := flags.Parse(&o)
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	var config steamroller.Config
	if o.ConfigPath.Path() != "" {
		var configBytes []byte
		configBytes, err = ioutil.ReadFile(o.ConfigPath.Path())
		if err != nil {
			log.Fatalf("Failed reading config file: %s", err)
		}
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			log.Fatalf("Failed unmarshaling config file: %s", err)
		}
	}

	pipelineBytes, err := ioutil.ReadFile(o.PipelinePath.Path())
	if err != nil {
		log.Fatalf("failed reading path: %s", err)
	}

	placeholderWrapper := yamlpatch.NewPlaceholderWrapper("{{", "}}")
	wrappedConfigBytes := placeholderWrapper.Wrap(pipelineBytes)

	var atcConfig atc.Config
	err = yaml.Unmarshal(wrappedConfigBytes, &atcConfig)
	if err != nil {
		log.Fatalf("failed unmarshaling config: %s", err)
	}

	steamrolledConfig, err := steamroller.Steamroll(config.ResourceMap, atcConfig)
	if err != nil {
		log.Fatalf("failed steamrolling config: %s", err)
	}

	bs, err := yaml.Marshal(steamrolledConfig)
	if err != nil {
		log.Fatalf("failed steamrolling config: %s", err)
	}

	unwrappedConfigBytes := placeholderWrapper.Unwrap(bs)

	f := bufio.NewWriter(os.Stdout)

	_, err = f.Write(unwrappedConfigBytes)
	if err != nil {
		log.Fatalf("failed to write steamrolled pipeline to stdout")
	}

	err = f.Flush()
	if err != nil {
		log.Fatalf("failed to flush stdout")
	}
}
