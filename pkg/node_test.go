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
		AfterEach(func() {
			clearResources()
		})
		When("all ok", func() {
			It("should summarize node info", func() {
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
				expectedNodeInfos := []nodeInfo{
					{
						nodeName:     "test1",
						devClassName: "ssd",
						used:         150,
						available:    100,
						usedRate:     60,
					},
					{
						nodeName:     "test1",
						devClassName: "hdd",
						used:         100,
						available:    200,
						usedRate:     33,
					},
					{
						nodeName:     "test1",
						devClassName: "00default",
						used:         0,
						available:    100,
						usedRate:     0,
					},
					{
						nodeName:     "test2",
						devClassName: "ssd",
						used:         100,
						available:    10,
						usedRate:     90,
					},
					{
						nodeName:     "test2",
						devClassName: "00default",
						used:         0,
						available:    10,
						usedRate:     0,
					},
				}
				nodeInfos, err := summarize(context.Background(), k8sClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(nodeInfos).Should(ConsistOf(expectedNodeInfos))
			})
		})
		When("node annotation var is zero", func() {
			It(`should output "used ratio" is zero`, func() {
				createNode("test1", map[string]string{
					"ssd":       "100",
					"00default": "100",
					"hdd":       "0",
				})
				createLv("lv1", "test1", "hdd", 0)
				expectedNodeInfos := []nodeInfo{
					{
						nodeName:     "test1",
						devClassName: "hdd",
						used:         0,
						available:    0,
						usedRate:     0,
					},
				}
				nodeInfos, err := summarize(context.Background(), k8sClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(nodeInfos).Should(HaveLen(3))
				Expect(nodeInfos).Should(ContainElements(expectedNodeInfos))
			})
		})
		When("device class is empty", func() {
			It(`should assign "00default"`, func() {
				createNode("test1", map[string]string{
					"00default": "100",
				})
				createLv("lv1", "test1", "", 100)
				expectedNodeInfos := []nodeInfo{
					{
						nodeName:     "test1",
						devClassName: "00default",
						used:         100,
						available:    100,
						usedRate:     50,
					},
				}
				nodeInfos, err := summarize(context.Background(), k8sClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(nodeInfos).Should(HaveLen(1))
				Expect(nodeInfos).Should(ContainElements(expectedNodeInfos))
			})
		})
		When("the node name which does not exist is used", func() {
			It(`should be ignored`, func() {
				createNode("test1", map[string]string{
					"ssd":       "100",
					"00default": "100",
				})
				createLv("lv1", "not_exist", "ssd", 100)
				expectedNodeInfos := []nodeInfo{
					{
						nodeName:     "test1",
						devClassName: "ssd",
						used:         0,
						available:    100,
						usedRate:     0,
					},
				}
				nodeInfos, err := summarize(context.Background(), k8sClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(nodeInfos).Should(HaveLen(2))
				Expect(nodeInfos).Should(ContainElements(expectedNodeInfos))
			})
		})
		When("the device class name which does not exist is used", func() {
			It(`should be ignored`, func() {
				createNode("test1", map[string]string{
					"ssd": "100",
				})
				createLv("lv1", "test1", "not_exist", 100)
				expectedNodeInfos := []nodeInfo{
					{
						nodeName:     "test1",
						devClassName: "ssd",
						used:         0,
						available:    100,
						usedRate:     0,
					},
				}
				nodeInfos, err := summarize(context.Background(), k8sClient)
				Expect(err).ToNot(HaveOccurred())
				Expect(nodeInfos).Should(HaveLen(1))
				Expect(nodeInfos).Should(ContainElements(expectedNodeInfos))
			})
		})
	})
	Context("summarize failure case", func() {
		When("node annotation var is string", func() {
			It("returns error", func() {
				createNode("test1", map[string]string{
					"ssd":       "aaa",
					"00default": "aaa",
				})
				createLv("lv1", "test1", "ssd", 50)
				_, err := summarize(context.Background(), k8sClient)
				Expect(err).To(HaveOccurred())
			})
		})
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

func clearResources() {
	err := k8sClient.DeleteAllOf(context.Background(), &corev1.Node{})
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.DeleteAllOf(context.Background(), &topolvmv1.LogicalVolume{})
	Expect(err).ToNot(HaveOccurred())
}
