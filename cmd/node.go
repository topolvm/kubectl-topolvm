package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topolvm/kubectl-topolvm/pkg"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Show information about TopoLVM logical volume for each node",
	Long:  "Show information about TopoLVM logical volume for each node",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pkg.PrintNodesInfo(cmd.Context(), kubeClient)
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
