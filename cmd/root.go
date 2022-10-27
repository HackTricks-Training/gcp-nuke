package cmd

import (
	"fmt"
	"sort"

	"github.com/dshelley66/gcp-nuke/pkg/config"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var (
		params  NukeParameters
		creds   gcputil.Credentials
		verbose bool
	)

	command := &cobra.Command{
		Use:   "gcp-nuke",
		Short: "gcp-nuke removes every resource from GCP",
		Long:  `A tool which removes every resource from a GCP project.  Use it with caution, since it cannot distinguish between production and non-production.`,
	}

	command.PreRun = func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.InfoLevel)
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		log.SetFormatter(&log.TextFormatter{
			EnvironmentOverrideColors: true,
		})
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		var err error

		err = params.Validate()
		if err != nil {
			return err
		}

		err = creds.Validate()
		if err != nil {
			return err
		}

		command.SilenceUsage = true

		config, err := config.Load(params.ConfigPath)
		if err != nil {
			log.Errorf("Failed to parse config file %s", params.ConfigPath)
			return err
		}

		n := NewNuke(params, &creds)

		n.Config = config
		n.Project = gcputil.NewProject(&creds)

		return n.Run()
	}

	command.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false,
		"Enables debug output.")

	command.PersistentFlags().StringVarP(
		&params.ConfigPath, "config", "c", "",
		"(required) Path to the nuke config file.")

	command.PersistentFlags().StringVarP(
		&creds.Keyfile, "keyfile", "k", "",
		"Path to file containing GCP service account credentials with read/write access to project to be nuked. "+
			"If not provided, authentication will be via application default credentials.")
	command.PersistentFlags().StringVarP(
		&creds.Project, "project", "p", "",
		"GCP Project to nuke")

	command.PersistentFlags().StringSliceVarP(
		&params.Targets, "target", "t", []string{},
		"Limit nuking to certain resource types (eg IAMServerCertificate). "+
			"This flag can be used multiple times.")
	command.PersistentFlags().StringSliceVarP(
		&params.Excludes, "exclude", "e", []string{},
		"Prevent nuking of certain resource types (eg IAMServerCertificate). "+
			"This flag can be used multiple times.")
	command.PersistentFlags().BoolVar(
		&params.NoDryRun, "no-dry-run", false,
		"If specified, it actually deletes found resources. "+
			"Otherwise it just lists all candidates.")
	command.PersistentFlags().BoolVar(
		&params.Force, "force", false,
		"Don't ask for confirmation before deleting resources. "+
			"Instead it waits 15s before continuing. Set --force-sleep to change the wait time.")
	command.PersistentFlags().IntVar(
		&params.ForceSleep, "force-sleep", 15,
		"If specified and --force is set, wait this many seconds before deleting resources. "+
			"Defaults to 15.")
	command.PersistentFlags().IntVar(
		&params.MaxWaitRetries, "max-wait-retries", 0,
		"If specified, the program will exit if resources are stuck in waiting for this many iterations. "+
			"0 (default) disables early exit.")
	command.PersistentFlags().BoolVarP(
		&params.Quiet, "quiet", "q", false,
		"Don't show filtered resources.")

	command.AddCommand(NewVersionCommand())
	command.AddCommand(NewResourceTypesCommand())

	return command
}

func NewResourceTypesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource-types",
		Short: "lists all available resource types",
		Run: func(cmd *cobra.Command, args []string) {
			names := resources.GetListerNames()
			sort.Strings(names)

			for _, resourceType := range names {
				fmt.Println(resourceType)
			}
		},
	}

	return cmd
}
