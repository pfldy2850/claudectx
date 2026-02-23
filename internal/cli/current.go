package cli

import (
	"fmt"

	"github.com/pfldy2850/claudectx/internal/context"
	"github.com/spf13/cobra"
)

func newCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the active context name",
		Args:  cobra.NoArgs,
		RunE:  runCurrent,
	}
}

func runCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	current, err := context.GetCurrent(cfg)
	if err != nil {
		return err
	}

	if current == "" {
		fmt.Println("No active context.")
		return nil
	}

	fmt.Println(current)
	return nil
}
