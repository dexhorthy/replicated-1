package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/replicatedhq/replicated/cli/print"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const (
	defaultYAMLDir = "manifests"
)

type kotsSingleSpec struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Content  string   `json:"content"`
	Children []string `json:"children"`
}

func (r *runners) InitReleaseCreate(parent *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new release",
		Long: `Create a new release by providing YAML configuration for the next release in
  your sequence.`,
		SilenceUsage: true,
	}

	parent.AddCommand(cmd)

	cmd.Flags().StringVar(&r.args.createReleaseYaml, "yaml", "", "The YAML config for this release. Use '-' to read from stdin.  Cannot be used with the `yaml-file` flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlFile, "yaml-file", "", "The file name with YAML config for this release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleaseYamlDir, "yaml-dir", "", "The directory containing multiple yamls for a Kots release.  Cannot be used with the `yaml` flag.")
	cmd.Flags().StringVar(&r.args.createReleasePromote, "promote", "", "Channel name or id to promote this release to")
	cmd.Flags().StringVar(&r.args.createReleasePromoteNotes, "release-notes", "", "When used with --promote <channel>, sets the **markdown** release notes")
	cmd.Flags().StringVar(&r.args.createReleasePromoteVersion, "version", "", "When used with --promote <channel>, sets the version label for the release in this channel")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteRequired, "required", false, "When used with --promote <channel>, marks this release as required during upgrades.")
	cmd.Flags().BoolVar(&r.args.createReleasePromoteEnsureChannel, "ensure-channel", false, "When used with --promote <channel>, will create the channel if it doesn't exist")
	cmd.Flags().BoolVar(&r.args.createReleaseAutoDefaults, "auto", false, "generate default values for use in CI")

	// not supported for KOTS
	cmd.Flags().MarkHidden("required")
	cmd.Flags().MarkHidden("yaml-file")
	cmd.Flags().MarkHidden("yaml")

	cmd.RunE = r.releaseCreate
	return nil
}

func (r *runners) gitSHABranch() (sha string, branch string, dirty bool, err error) {
	path := "."
	rev := "HEAD"
	repository, err := git.PlainOpen(path)
	if err != nil {
		return "", "", false, errors.Wrapf(err, "open %q", path)
	}
	h, err := repository.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", "", false, errors.Wrapf(err, "resolve revision")
	}
	head, err := repository.Head()
	if err != nil {
		return "", "", false, errors.Wrapf(err, "resolve HEAD")
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return "", "", false, errors.Wrap(err, "get git worktree")
	}
	status, err := worktree.Status()
	if err != nil {
		return "", "", false, errors.Wrap(err, "get git status")
	}

	return h.String()[0:6], head.Name().Short(), !status.IsClean(), nil
}

func (r *runners) setKOTSDefaults() error {
	rev, branch, isDirty, err := r.gitSHABranch()
	if err != nil {
		return errors.Wrapf(err, "get git properties")
	}

	if r.args.createReleaseYamlDir == "" {
		r.args.createReleaseYamlDir = "./manifests"
	}

	if r.args.createReleasePromoteNotes == "" {
		r.args.createReleasePromoteNotes = fmt.Sprintf(
			`CLI release by %s on %s`, os.Getenv("USER"), time.Now().Format(time.RFC822))
	}

	if r.args.createReleasePromote == "" {
		r.args.createReleasePromote = branch
		if branch == "master" || branch == "main" {
			r.args.createReleasePromote = "Unstable"
		}
	}

	if r.args.createReleasePromoteVersion == "" {
		dirtyStatus := ""
		if isDirty {
			dirtyStatus = "-dirty"
		}
		r.args.createReleasePromoteVersion = fmt.Sprintf("%s-%s%s", r.args.createReleasePromote, rev, dirtyStatus)
	}

	r.args.createReleasePromoteEnsureChannel = true

	return nil
}

func (r *runners) releaseCreate(_ *cobra.Command, _ []string) error {

	log := print.NewLogger(r.w)

	if r.appType == "kots" && r.args.createReleaseAutoDefaults {
		log.ActionWithSpinner("Reading Environment")
		err := r.setKOTSDefaults()
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "resolve kots defaults")
		}
		time.Sleep(500 * time.Millisecond)
		log.FinishSpinner()

		fmt.Fprintf(r.w, `
Prepared to create release with defaults:

    yaml-dir        %q
    promote         %q
    version         %q
    release-notes   %q
    ensure-channel  %t

`, r.args.createReleaseYamlDir, r.args.createReleasePromote, r.args.createReleasePromoteVersion, r.args.createReleasePromoteNotes, r.args.createReleasePromoteEnsureChannel)
		confirmed, err := promptForConfirm()
		if err != nil {
			return err
		}
		if confirmed != "y" {
			return errors.New("configuration declined")
		}

	}

	if r.args.createReleaseYaml == "" && r.args.createReleaseYamlFile == "" && r.appType != "kots" {
		return errors.New("one of --yaml, --yaml-file must be provided")
	}

	if r.args.createReleaseYaml != "" && r.args.createReleaseYamlFile != "" {
		return errors.New("only one of --yaml or --yaml-file may be specified")
	}

	if r.args.createReleaseYamlDir == "" && r.appType == "kots" {
		return errors.New("--yaml-dir flag must be provided for KOTS applications")
	}

	// can't ensure a channel if you didn't pass one
	if r.args.createReleasePromoteEnsureChannel && r.args.createReleasePromote == "" {
		return errors.New("cannot use the flag --ensure-channel without also using --promote <channel> ")
	}

	// we check this again below, but lets be explicit and fail fast
	if r.args.createReleasePromoteEnsureChannel && !(r.appType == "ship" || r.appType == "kots") {
		return errors.Errorf("the flag --ensure-channel is only supported for KOTS and Ship applications, app %q is of type %q", r.appID, r.appType)
	}

	if r.args.createReleasePromoteRequired && r.appType == "kots" {
		return errors.Errorf("the --required flag is not supported for KOTS applications")
	}

	if r.args.createReleaseYamlFile != "" && r.appType == "kots" {
		return errors.Errorf("the --yaml-file flag is not supported for KOTS applications, use --yaml-dir instead")
	}

	if r.args.createReleaseYaml != "" && r.appType == "kots" {
		return errors.Errorf("the --yaml flag is not supported for KOTS applications, use --yaml-dir instead")
	}

	if r.args.createReleaseYaml == "-" {
		bytes, err := ioutil.ReadAll(r.stdin)
		if err != nil {
			return errors.Wrap(err, "read from stdin")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlFile != "" {
		bytes, err := ioutil.ReadFile(r.args.createReleaseYamlFile)
		if err != nil {
			return errors.Wrap(err, "read file yaml")
		}
		r.args.createReleaseYaml = string(bytes)
	}

	if r.args.createReleaseYamlDir != "" {
		fmt.Fprintln(r.w)
		log.ActionWithSpinner("Reading manifests from %s", r.args.createReleaseYamlDir)
		var err error
		r.args.createReleaseYaml, err = readYAMLDir(r.args.createReleaseYamlDir)
		if err != nil {
			log.FinishSpinnerWithError()
			return errors.Wrap(err, "read yaml dir")
		}
		log.FinishSpinner()
	}

	// if the --promote param was used make sure it identifies exactly one
	// channel before proceeding
	var promoteChanID string
	if r.args.createReleasePromote != "" {
		var err error
		promoteChanID, err = r.getOrCreateChannelForPromotion(
			r.args.createReleasePromote,
			r.args.createReleasePromoteEnsureChannel,
		)
		if err != nil {
			return errors.Wrapf(err, "get or create channel %q for promotion", promoteChanID)
		}
	}

	log.ActionWithSpinner("Creating Release")
	release, err := r.api.CreateRelease(r.appID, r.appType, r.args.createReleaseYaml)
	if err != nil {
		log.FinishSpinnerWithError()
		return err
	}
	log.FinishSpinner()

	log.ChildActionWithoutSpinner("SEQUENCE: %d", release.Sequence)

	if promoteChanID != "" {
		log.ActionWithSpinner("Promoting")
		if err := r.api.PromoteRelease(
			r.appID,
			r.appType,
			release.Sequence,
			r.args.createReleasePromoteVersion,
			r.args.createReleasePromoteNotes,
			r.args.createReleasePromoteRequired,
			promoteChanID,
		); err != nil {
			log.FinishSpinnerWithError()
			return err
		}
		log.FinishSpinner()

		// ignore error since operation was successful
		log.ChildActionWithoutSpinner("Channel %s successfully set to release %d\n", promoteChanID, release.Sequence)
	}

	return nil
}

func (r *runners) getOrCreateChannelForPromotion(channelName string, createIfAbsent bool) (string, error) {

	description := "" // todo: do we want a flag for the desired channel description

	channel, err := r.api.GetChannelByName(
		r.appID,
		r.appType,
		channelName,
		description,
		createIfAbsent,
	)
	if err != nil {
		return "", errors.Wrapf(err, "get-or-create channel %q", channelName)
	}

	return channel.ID, nil
}

func encodeKotsFile(prefix, path string, info os.FileInfo, err error) (*kotsSingleSpec, error) {
	if err != nil {
		return nil, err
	}

	singlefile := strings.TrimPrefix(filepath.Clean(path), filepath.Clean(prefix)+"/")

	if info.IsDir() {
		return nil, nil
	}
	if strings.HasPrefix(info.Name(), ".") {
		return nil, nil
	}
	ext := filepath.Ext(info.Name())
	switch ext {
	case ".tgz", ".gz", ".yaml", ".yml":
		// continue
	default:
		return nil, nil
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", path)
	}

	var str string
	switch ext {
	case ".tgz", ".gz":
		str = base64.StdEncoding.EncodeToString(bytes)
	default:
		str = string(bytes)
	}

	return &kotsSingleSpec{
		Name:     info.Name(),
		Path:     singlefile,
		Content:  str,
		Children: []string{},
	}, nil
}

func readYAMLDir(yamlDir string) (string, error) {

	var allKotsReleaseSpecs []kotsSingleSpec
	err := filepath.Walk(yamlDir, func(path string, info os.FileInfo, err error) error {
		spec, err := encodeKotsFile(yamlDir, path, info, err)
		if err != nil {
			return err
		} else if spec == nil {
			return nil
		}
		allKotsReleaseSpecs = append(allKotsReleaseSpecs, *spec)
		return nil
	})
	if err != nil {
		return "", errors.Wrapf(err, "walk %s", yamlDir)
	}

	jsonAllYamls, err := json.Marshal(allKotsReleaseSpecs)
	if err != nil {
		return "", errors.Wrap(err, "marshal spec")
	}
	return string(jsonAllYamls), nil
}

func promptForConfirm() (string, error) {

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . | bold }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "Create release with these properties? [Y/n]",
		Templates: templates,
		Default:   "y",
		Validate: func(input string) error {
			if input != "y" && input != "n" {
				return errors.New(`please choose "y" or "n"`)
			}

			return nil
		},
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
