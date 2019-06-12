package util

import (
	"fmt"
	"github.com/pkg/errors"
	kotskinds "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/kots/kotskinds/multitype"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	sigyaml "sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

type HelmConverter interface {
	ConvertValues(helmValues string, chartYAML string) (*FakeHelmChart, *kotskinds.Config, error)
}

// I can't for the life of me get the kotskinds one to serialize nicely
type FakeHelmChart struct {
	kotskinds.HelmChart `json:",inline"`
	Spec                FakeHelmChartSpec `json:"spec"`
}

// I can't for the life of me get the kotskinds one to serialize nicely
type FakeHelmChartSpec struct {
	kotskinds.HelmChartSpec `json:",inline"`
	Values                  map[string]interface{} `json:"values"`
}

// while we're at it lets fake this too
type FakeChartYAML struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

func NewHelmConverter() HelmConverter {
	return &helmConverter{}
}

var _ HelmConverter = &helmConverter{}

type helmConverter struct {
}

func (c helmConverter) ConvertValues(valuesYAML string, chartYAML string) (*FakeHelmChart, *kotskinds.Config, error) {
	var values yaml.MapSlice
	err := yaml.Unmarshal([]byte(valuesYAML), &values)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unmarshal values")
	}

	var path []string

	helmValuesWithConfigOption, configItems := c.convertValuesRec(values, path)

	config := &kotskinds.Config{
		Spec: kotskinds.ConfigSpec{
			Groups: []kotskinds.ConfigGroup{
				{
					Title: "Generated",
					Items: configItems,
				},
			},
		},
	}
	var chart FakeChartYAML
	err = sigyaml.Unmarshal([]byte(chartYAML), &chart)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unmarshal chart yaml")
	}

	helmChart := &FakeHelmChart{
		HelmChart: kotskinds.HelmChart{
			ObjectMeta: v1.ObjectMeta{
				Name: chart.Name,
			},
			TypeMeta: v1.TypeMeta{
				Kind:       "HelmChart",
				APIVersion: "kots.io/v1beta1",
			},
		},
		Spec: FakeHelmChartSpec{
			HelmChartSpec: kotskinds.HelmChartSpec{
				Chart: kotskinds.ChartIdentifier{
					Name:         chart.Name,
					ChartVersion: chart.Version,
				},
			},
			Values: helmValuesWithConfigOption,
		},
	}

	return helmChart, config, nil
}

func (c helmConverter) convertValuesRec(in yaml.MapSlice, path []string) (map[string]interface{}, []kotskinds.ConfigItem) {
	valuesYAMLAcc := map[string]interface{}{}
	var configItems []kotskinds.ConfigItem

	for _, item := range in {
		key, ok := item.Key.(string)
		if !ok {
			// skip non-string keys (log me?)
			continue
		}

		newPath := append(path, key)

		configItemName := strings.Join(newPath, ".")
		configItemTitle := configItemName

		appendScalar := func() {
			valuesYAMLAcc[key] = fmt.Sprintf(
				"{{repl ConfigOption %q }}",
				configItemName,
			)
		}

		// todo support for more types
		switch typedValue := item.Value.(type) {
		case int:
			appendScalar()
			configItems = append(configItems, kotskinds.ConfigItem{
				Name:    configItemName,
				Title:   configItemTitle,
				Default: multitype.FromString(strconv.Itoa(typedValue)),
				Type:    "text",
			})
		case string:
			appendScalar()
			configItems = append(configItems, kotskinds.ConfigItem{
				Name:    configItemName,
				Title:   configItemTitle,
				Default: multitype.FromString(typedValue),
				Type:    "text",
			})
		case bool:
			appendScalar()
			configItems = append(configItems, kotskinds.ConfigItem{
				Name:    configItemName,
				Title:   configItemTitle,
				Default: multitype.FromString(fmt.Sprintf("%v", typedValue)),
				Type:    "bool",
			})
		case yaml.MapSlice:
			value, items := c.convertValuesRec(typedValue, newPath)
			valuesYAMLAcc[key] = value
			configItems = append(configItems, items...)
		case []interface{}:
		default:
			if typedValue == nil {
				appendScalar()
				configItems = append(configItems, kotskinds.ConfigItem{
					Name:    configItemName,
					Title:   configItemTitle,
					Default: multitype.FromString(fmt.Sprintf("%v", typedValue)),
					Type:    "text",
				})
			} else {
				// todo need a real logger here
				fmt.Fprint(os.Stderr, fmt.Sprintf("Unsupported value type \"%T\" at %q: %q, using default.\n", configItemName, typedValue))
				valuesYAMLAcc[key] = item.Value
			}
		}
	}

	return valuesYAMLAcc, configItems
}
