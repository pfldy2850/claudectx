package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pfldy2850/claudectx/internal/context"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a saved context",
		Args:    cobra.ExactArgs(1),
		RunE:    runDelete,
	}
}

func runDelete(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	slug := context.Slugify(args[0])

	if !context.ContextExists(cfg.ContextsDir(), slug) {
		return fmt.Errorf("context %q not found", slug)
	}

	// Prevent deleting the active context
	current, _ := context.GetCurrent(cfg)
	if current == slug {
		return fmt.Errorf("cannot delete active context %q; switch to another context first", slug)
	}

	if !force {
		fmt.Printf("Delete context %q? [y/N] ", slug)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if dryRun {
		fmt.Printf("[dry-run] Would delete context %q\n", slug)
		return nil
	}

	if err := context.DeleteContext(cfg.ContextsDir(), slug); err != nil {
		return err
	}

	fmt.Printf("Context %q deleted.\n", slug)
	return nil
}
