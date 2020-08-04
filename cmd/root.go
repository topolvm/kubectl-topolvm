package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	topolvmv1 "github.com/topolvm/topolvm/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var kubeClient client.Client

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl-topolvm",
	Short: "the utility command for TopoLVM",
	Long:  "the utility command for TopoLVM",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		kubeConfigFlags := genericclioptions.NewConfigFlags(true)
		kubeConfigFlags.AddFlags(cmd.PersistentFlags())

		factory := util.NewFactory(util.NewMatchVersionFlags(kubeConfigFlags))
		config, err := factory.ToRESTConfig()
		if err != nil {
			return err
		}

		scheme := runtime.NewScheme()
		err = clientgoscheme.AddToScheme(scheme)
		if err != nil {
			return err
		}

		err = topolvmv1.AddToScheme(scheme)
		if err != nil {
			return err
		}

		kubeClient, err = client.New(config, client.Options{Scheme: scheme})
		if err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
