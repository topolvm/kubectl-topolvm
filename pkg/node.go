package pkg

import (
	"context"
	"fmt"
	"github.com/topolvm/topolvm"
	topolvmv1 "github.com/topolvm/topolvm/api/v1"
	corev1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"text/tabwriter"
)

type nodeInfo struct {
	nodeName     string
	devClassName string
	used         int64
	available    int64
	usedRate     int64
}

func PrintNodesInfo(ctx context.Context, kubeClient client.Client) error {
	nodeInfos, err := summarize(ctx, kubeClient)
	if err != nil {
		return err
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	_, err = fmt.Fprintln(w, "NODE\tDEVICECLASS\tUSED\tAVAILABLE\tUSE%")
	if err != nil {
		return err
	}
	for _, info := range nodeInfos {
		_, err = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d%%\n", info.nodeName, info.devClassName, info.used, info.available, info.usedRate)
		if err != nil {
			return err
		}
	}
	return w.Flush()
}

func summarize(ctx context.Context, kubeClient client.Client) ([]nodeInfo, error) {
	var nodes corev1.NodeList
	err := kubeClient.List(ctx, &nodes, &client.ListOptions{})
	if err != nil {
		return nil, err
	}

	var lvs topolvmv1.LogicalVolumeList
	err = kubeClient.List(ctx, &lvs, &client.ListOptions{})
	if err != nil {
		return nil, err
	}

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

	nodeInfos := make([]nodeInfo, 0)
	for _, node := range nodes.Items {
		nodeName := node.Name
		for key, val := range node.Annotations {
			if !strings.HasPrefix(key, topolvm.CapacityKeyPrefix) {
				continue
			}
			devClassName := strings.Split(key, "/")[1]
			usedRate := int64(0)
			available, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, err
			}
			used := sizeMap[nodeName][devClassName]
			capacity := available + used
			if capacity != 0 {
				usedRate = used * 100 / capacity
			}
			info := nodeInfo{
				nodeName:     nodeName,
				devClassName: devClassName,
				used:         used,
				available:    available,
				usedRate:     usedRate,
			}
			nodeInfos = append(nodeInfos, info)
		}
	}
	return nodeInfos, nil
}
