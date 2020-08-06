package pkg

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topolvm/topolvm"
	topolvmv1 "github.com/topolvm/topolvm/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("node command test", func() {
	Context("summarize success case", func() {
		It("create test data", func() {
			createNode("test1", map[string]string{
				"ssd":       "100",
				"00default": "100",
				"hdd":       "200",
			})
			createNode("test2", map[string]string{
				"ssd":       "10",
				"00default": "10",
			})
			createLv("lv1", "test1", "ssd", 50)
			createLv("lv2", "test1", "ssd", 100)
			createLv("lv3", "test1", "hdd", 100)
			createLv("lv4", "test2", "ssd", 50)
			createLv("lv5", "test2", "ssd", 50)
		})
		It("should summarize node info", func() {
			expectedNodeInfos := []nodeInfo{
				{
					nodeName:     "test1",
					devClassName: "ssd",
					used:         150,
					available:    100,
					usedRate:     60,
				},
			}
			nodeInfos, err := summarize(context.Background(), k8sClient)
			Expect(err).ToNot(HaveOccurred())
			Expect(nodeInfos).Should(HaveLen(5))
			Expect(nodeInfos).Should(ContainElements(expectedNodeInfos))
		})
	})
	Context("confirming summarize()", func() {

	})
})

func createNode(name string, devClasses map[string]string) {
	node := corev1.Node{}
	node.Name = name

	node.Annotations = make(map[string]string)
	for k, v := range devClasses {
		node.Annotations[topolvm.CapacityKeyPrefix+k] = v
	}
	err := k8sClient.Create(context.Background(), &node)
	Expect(err).ToNot(HaveOccurred())
}

func createLv(name, nodeName, devClass string, size int64) {
	lv := topolvmv1.LogicalVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: topolvmv1.LogicalVolumeSpec{
			Name:        name,
			NodeName:    nodeName,
			Size:        *resource.NewQuantity(1, resource.BinarySI),
			DeviceClass: devClass,
		},
		Status: topolvmv1.LogicalVolumeStatus{
			CurrentSize: resource.NewQuantity(size, resource.BinarySI),
		},
	}
	ctx := context.Background()
	err := k8sClient.Create(ctx, &lv)
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.Status().Update(ctx, &lv)
	Expect(err).ToNot(HaveOccurred())
}
