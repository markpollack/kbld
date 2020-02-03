package cmd

import (
	"fmt"

	"github.com/cppforlife/go-cli-ui/ui"
	cmdcore "github.com/k14s/kbld/pkg/kbld/cmd/core"
	"github.com/spf13/cobra"
)

const (
	Version = "0.13.0"
)

type VersionOptions struct {
	ui ui.UI
}

func NewVersionOptions(ui ui.UI) *VersionOptions {
	return &VersionOptions{ui}
}

func NewVersionCmd(o *VersionOptions, flagsFactory cmdcore.FlagsFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print client version",
		RunE:  func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	return cmd
}

func (o *VersionOptions) Run() error {
	o.ui.PrintBlock([]byte(fmt.Sprintf("kbld version %s\n", Version)))

	return nil
}
