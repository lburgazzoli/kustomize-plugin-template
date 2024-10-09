package main

import (
	"os"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

	ktpl "github.com/lburgazzoli/kustomize-plugin-template/pkg/krm/template"
)

func main() {
	c := ktpl.Configuration{}

	p := framework.SimpleProcessor{
		Config: &c,
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			f := ktpl.Function{
				Spec: c.Spec,
			}

			return f.Apply(nodes)
		}),
	}

	cmd := command.Build(p, command.StandaloneDisabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
