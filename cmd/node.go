package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/topolvm/topolvm"
	topolvmv1 "github.com/topolvm/topolvm/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Show information about TopoLVM logical volume for each node",
	Long:  "Show information about TopoLVM logical volume for each node",
	RunE: func(cmd *cobra.Command, args []string) error {

		var nodes corev1.NodeList
		err := kubeClient.List(cmd.Context(), &nodes, &client.ListOptions{})
		if err != nil {
			return err
		}

		var lvs topolvmv1.LogicalVolumeList
		err = kubeClient.List(cmd.Context(), &lvs, &client.ListOptions{})
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
}
