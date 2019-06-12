package util

import (
	"github.com/replicatedhq/libyaml"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestHelmBuildValues(t *testing.T) {
	tests := []struct {
		name        string
		values      string
		expect      string
		expectItems []*libyaml.ConfigItem
	}{
		{
			name:        "empty",
			values:      "",
			expect:      "{}",
			expectItems: []*libyaml.ConfigItem{},
		},
		{
			name:        "empty 2",
			values:      "{}",
			expect:      "{}",
			expectItems: []*libyaml.ConfigItem{},
		},
		{
			name:   "one value",
			values: `reticulate_splines: please`,
			expect: `reticulate_splines: '{{repl ConfigOption "reticulate_splines" }}'`,
			expectItems: []*libyaml.ConfigItem{
				{
					Title:   "reticulate_splines",
					Name:    "reticulate_splines",
					Default: "please",
					Type:    "text",
				},
			},
		},
		{
			name: "two values",
			values: `
reticulate_splines: please
deploy_adjunct_frombulator: false
`,
			expect: `
reticulate_splines: '{{repl ConfigOption "reticulate_splines" }}'
deploy_adjunct_frombulator: '{{repl ConfigOption "deploy_adjunct_frombulator" }}'
`,
			expectItems: []*libyaml.ConfigItem{
				{
					Title:   "reticulate_splines",
					Name:    "reticulate_splines",
					Default: "please",
					Type:    "text",
				},
				{
					Title:   "deploy_adjunct_frombulator",
					Name:    "deploy_adjunct_frombulator",
					Default: "false",
					Type:    "bool",
				},
			},
		},
		{
			name: "one nested value",
			values: `
reticulate_splines:
  depth: 4
`,
			expect: `
reticulate_splines:
  depth: '{{repl ConfigOption "reticulate_splines.depth" }}'
`,
			expectItems: []*libyaml.ConfigItem{
				{
					Title:   "reticulate_splines.depth",
					Name:    "reticulate_splines.depth",
					Default: "4",
					Type:    "text",
				},
			},
		},
		{
			name: "multiple nested values",
			values: `
reticulate_splines:
  depth: 4
persistence:
  postgres:
    enable: false
  redis:
    enable: false
    replicas: 2
`,
			expect: `
reticulate_splines:
  depth: '{{repl ConfigOption "reticulate_splines.depth" }}'
persistence:
  postgres:
    enable: '{{repl ConfigOption "persistence.postgres.enable" }}'
  redis:
    enable: '{{repl ConfigOption "persistence.redis.enable" }}'
    replicas: '{{repl ConfigOption "persistence.redis.replicas" }}'
`,
			expectItems: []*libyaml.ConfigItem{
				{
					Title:   "reticulate_splines.depth",
					Name:    "reticulate_splines.depth",
					Default: "4",
					Type:    "text",
				},
				{
					Title:   "persistence.postgres.enable",
					Name:    "persistence.postgres.enable",
					Default: "false",
					Type:    "bool",
				},
				{
					Title:   "persistence.redis.enable",
					Name:    "persistence.redis.enable",
					Default: "false",
					Type:    "bool",
				},
				{
					Title:   "persistence.redis.replicas",
					Name:    "persistence.redis.replicas",
					Default: "2",
					Type:    "text",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			hc := &helmConverter{}
			actual, group, err := hc.ConvertValues(test.values)

			req.NoError(err, "convert helm values")
			req.Equal(strings.TrimSpace(test.expect), strings.TrimSpace(actual))
			for i, item := range group[0].Items {
				req.True(len(test.expectItems) > i, "received unexpected item %d: %s", i, item.Name)
				req.Equal(test.expectItems[i], item)
			}
		})
	}
}
