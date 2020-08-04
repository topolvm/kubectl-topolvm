package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/topolvm/topolvm"
	topolvmv1 "github.com/topolvm/topolvm/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		kubeClient, err := client.New(config, client.Options{Scheme: scheme})
		if err != nil {
			return err
		}

		var nodes corev1.NodeList
		err = kubeClient.List(context.Background(), &nodes, &client.ListOptions{})
		if err != nil {
			return err
		}

		var lvs topolvmv1.LogicalVolumeList
		err = kubeClient.List(context.Background(), &lvs, &client.ListOptions{})
		if err != nil {
			return err
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(w, "NODE\tDEVICECLASS\tUSED\tAVAILABLE\tUSE%")

		sizeMap := make(map[string]map[string]int64)

		for _, lv := range lvs.Items {
			nodeName := lv.Spec.NodeName
			deviceClass := lv.Spec.DeviceClass
			currentSize := lv.Status.CurrentSize

			if _, ok := sizeMap[nodeName]; !ok {
				sizeMap[nodeName] = make(map[string]int64)
			}

			if deviceClass == "" {
				deviceClass = topolvm.DefaultDeviceClassAnnotationName
			}

			sizeMap[nodeName][deviceClass] += currentSize.Value()
		}

		for _, node := range nodes.Items {
			nodeName := node.Name
			for key, val := range node.Annotations {
				if strings.HasPrefix(key, topolvm.CapacityKeyPrefix) {
					devClassName := strings.Split(key, "/")[1]
					usedRate := int64(0)
					available, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						return err
					}
					used := sizeMap[nodeName][devClassName]
					capacity := available + used
					if capacity != 0 {
						usedRate = used * 100 / capacity
					}

					fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d%%\n", nodeName, devClassName, used, available, usedRate)
				}
			}
		}

		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
