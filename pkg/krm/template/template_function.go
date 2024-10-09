package template

import (
	"fmt"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"
	"text/template"
)

type Function struct {
	Spec ConfigurationSpec
}

func (p *Function) Apply(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	w := strings.Builder{}

	for i := range nodes {
		data, err := nodes[i].String()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal node to yaml: %w", err)
		}

		tmpl, err := template.New(nodes[i].GetName()).Delims("<<", ">>").Parse(data)
		if err != nil {
			return nil, fmt.Errorf("error parsing template %q: %v", nodes[i].GetName(), err)
		}

		w.Reset()

		err = tmpl.Execute(&w, p.Spec.Values)
		if err != nil {
			return nil, fmt.Errorf("error executing template: %v", err)
		}

		ret, err := yaml.Parse(w.String())
		if err != nil {
			return nil, fmt.Errorf("unable to map target %v: %w", data, err)
		}

		nodes[i] = ret
	}

	return nodes, nil
}
