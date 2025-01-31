package sshkey

import (
	"github.com/hetznercloud/cli/internal/cmd/output"
	"github.com/hetznercloud/cli/internal/cmd/util"
	"github.com/hetznercloud/cli/internal/state"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/hetznercloud/hcloud-go/hcloud/schema"
	"github.com/spf13/cobra"
)

var listTableOutput *output.Table

func init() {
	listTableOutput = output.NewTable().
		AddAllowedFields(hcloud.SSHKey{}).
		AddFieldFn("labels", output.FieldFn(func(obj interface{}) string {
			sshKey := obj.(*hcloud.SSHKey)
			return util.LabelsToString(sshKey.Labels)
		})).
		AddFieldFn("created", output.FieldFn(func(obj interface{}) string {
			sshKey := obj.(*hcloud.SSHKey)
			return util.Datetime(sshKey.Created)
		}))
}

func newListCommand(cli *state.State) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [FLAGS]",
		Short: "List SSH keys",
		Long: util.ListLongDescription(
			"Displays a list of SSH keys.",
			listTableOutput.Columns(),
		),
		TraverseChildren:      true,
		DisableFlagsInUseLine: true,
		PreRunE:               cli.EnsureToken,
		RunE:                  cli.Wrap(runList),
	}
	output.AddFlag(cmd, output.OptionNoHeader(), output.OptionColumns(listTableOutput.Columns()), output.OptionJSON())
	cmd.Flags().StringP("selector", "l", "", "Selector to filter by labels")
	return cmd
}

func runList(cli *state.State, cmd *cobra.Command, args []string) error {
	outOpts := output.FlagsForCommand(cmd)

	labelSelector, _ := cmd.Flags().GetString("selector")
	opts := hcloud.SSHKeyListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: labelSelector,
			PerPage:       50,
		},
	}
	sshKeys, err := cli.Client().SSHKey.AllWithOpts(cli.Context, opts)
	if err != nil {
		return err
	}

	if outOpts.IsSet("json") {
		var sshKeySchemas []schema.SSHKey
		for _, sshKey := range sshKeys {
			sshKeySchema := schema.SSHKey{
				ID:          sshKey.ID,
				Name:        sshKey.Name,
				Fingerprint: sshKey.Fingerprint,
				PublicKey:   sshKey.PublicKey,
				Labels:      sshKey.Labels,
				Created:     sshKey.Created,
			}
			sshKeySchemas = append(sshKeySchemas, sshKeySchema)
		}
		return util.DescribeJSON(sshKeySchemas)
	}

	cols := []string{"id", "name", "fingerprint"}
	if outOpts.IsSet("columns") {
		cols = outOpts["columns"]
	}

	tw := listTableOutput
	if err = tw.ValidateColumns(cols); err != nil {
		return err
	}

	if !outOpts.IsSet("noheader") {
		tw.WriteHeader(cols)
	}
	for _, sshKey := range sshKeys {
		tw.Write(cols, sshKey)
	}
	tw.Flush()
	return nil
}
