package template_test

import (
	"errors"
	"github.com/lburgazzoli/kustomize-plugin-template/pkg/krm/template"
	"io"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"testing"

	. "github.com/onsi/gomega"

	gyq "github.com/lburgazzoli/gomega-matchers/pkg/matchers/yq"
)

const c = `
apiVersion: kustomize.lburgazzoli.github.io/v1alpha1
kind: TemplateTransform
metadata:
  name: template-transformer
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/lburgazzoli/kustomize-plugin-template:latest
spec:
  values:
    resources:
      type: fixed
      fixed:
        replicas: 1
        resources:
          limits:
            cpu: 123m
            memory: 456Mi
          requests:
            cpu: 321m
            memory: 654Mi
`

const v = `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
spec:
  replicas: << .resources.fixed.replicas >>
  selector:
    matchLabels:
      control-plane: foo-component
  template:
    metadata:
      labels:
        app: foo-component
    spec:
      containers:
        - name: manager
          image: quay.io/lburgazzoli/component:latest
          resources:
            limits:
              cpu: '<< .resources.fixed.resources.limits.cpu >>'
              memory: '<< .resources.fixed.resources.limits.memory >>'
            requests:
              cpu: '<< .resources.fixed.resources.requests.cpu >>'
              memory: '<< .resources.fixed.resources.requests.memory >>'
`

func TestTemplate(t *testing.T) {
	t.Parallel()

	g := NewWithT(t)
	cfg := template.Configuration{}

	p := framework.SimpleProcessor{
		Config: &cfg,
		Filter: kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			f := template.Function{
				Spec: cfg.Spec,
			}

			return f.Apply(nodes)
		}),
	}

	w := DocumentSplitter{}

	rw := &kio.ByteReadWriter{
		Writer:                &w,
		KeepReaderAnnotations: false,
		NoWrap:                true,
		FunctionConfig:        yaml.MustParse(c),
		Reader:                strings.NewReader(v),
	}

	err := framework.Execute(p, rw)
	g.Expect(err).ToNot(HaveOccurred())

	items, err := w.Items()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(items).To(HaveLen(1))

	g.Expect(items[0]).Should(
		WithTransform(gyq.Extract(`.spec.template.spec.containers[0].resources`),
			And(
				gyq.Match(`.limits.cpu == "123m"`),
				gyq.Match(`.limits.memory == "456Mi"`),
				gyq.Match(`.requests.cpu == "321m"`),
				gyq.Match(`.requests.memory == "654Mi"`),
			),
		),
	)
	g.Expect(items[0]).Should(
		WithTransform(gyq.Extract(`.spec`),
			And(
				gyq.Match(`.replicas == "1"`),
			),
		),
	)
}

type DocumentSplitter struct {
	buffer strings.Builder
}

func (in *DocumentSplitter) Write(p []byte) (int, error) {
	return in.buffer.Write(p)
}

func (in *DocumentSplitter) Reset() {
	in.buffer.Reset()
}

func (in *DocumentSplitter) Items() ([]string, error) {
	items := make([]string, 0)

	r := strings.NewReader(in.buffer.String())
	dec := yaml.NewDecoder(r)

	for {
		var node yaml.Node

		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		item, err := yaml.Marshal(&node)
		if err != nil {
			return nil, err
		}

		items = append(items, string(item))
	}

	return items, nil
}
