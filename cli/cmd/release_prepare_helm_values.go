package cmd

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	kotskinds "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/replicated/pkg/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func (r *runners) InitPrepareHelmValues(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "prepare-helm-values [PATH]",
		Short: "Convert a helm values.yaml to be used with a replicated release",
		Long: `Convert a helm values.yaml to be used with a replicated release. 

PATH is optional, will default to reading a values file from a 
"values.yaml" in the current working directory.

`,
	}

	parent.AddCommand(cmd)
	cmd.RunE = r.prepareHelmValues
}

func (r *runners) prepareHelmValues(cmd *cobra.Command, args []string) error {
	chartPath := ""
	if len(args) == 1 {
		chartPath = args[0]
	}

	valuesBytes, err := ioutil.ReadFile(filepath.Join(chartPath, "values.yaml"))
	if err != nil {
		return errors.Wrap(err, "read values file %q")
	}

	chartYAMLBytes, err := ioutil.ReadFile(filepath.Join(chartPath, "Chart.yaml"))
	if err != nil {
		return errors.Wrap(err, "read values file %q")
	}

	var helmChart *util.FakeHelmChart
	var config *kotskinds.Config
	helmChart, config, err = r.helmConverter.ConvertValues(string(valuesBytes), string(chartYAMLBytes))
	if err != nil {
		return errors.Wrap(err, "convert Helm Values")
	}

	helmChartYAML, err := yaml.Marshal(helmChart)
	if err != nil {
		return errors.Wrap(err, "marshal helm chart YAML")
	}

	configYAML, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshal config YAML")
	}

	writeChart, err := promptForPath( "HelmChart destination: ", fmt.Sprintf("manifests/%s.yaml", helmChart.Spec.Chart.Name), dirExists)
	writeConfig, err := promptForPath("Config destination:    ", "manifests/config.yaml", dirExists)

	err = ioutil.WriteFile(writeChart, helmChartYAML, 400)
	if err != nil {
		return errors.Wrapf(err, "write %q", writeChart)
	}
	err = ioutil.WriteFile(writeConfig, configYAML, 400)
	if err != nil {
		return errors.Wrapf(err, "write %q", writeConfig)
	}
	return nil
}

func dirExists(input string) error {
	dir := filepath.Dir(input)
	if _, err := os.Stat(dir); err != nil {
		return err
	}
	return nil
}

func promptForPath(promptText, defaultPath string, validate func(string) error) (string, error) {

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | bold }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     promptText,
		Templates: templates,
		Default:   defaultPath,
		Validate: validate,
	}

	for {
		result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(-1)
			}
			continue
		}

		return result, nil
	}
}
