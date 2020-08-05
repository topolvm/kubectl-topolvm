K8S_VER=1.18.6
ETCD_VER=3.3.22
TOPOLVM_VER=0.5.2
TESTBIN_DIR=bin
CRD_DIR=crd
TOPOLVM_CRD=$(CRD_DIR)/topolvm.cybozu.com_logicalvolumes.yaml

test: $(TOPOLVM_CRD)
	KUBEBUILDER_ASSETS=$(CURDIR)/$(TESTBIN_DIR) go test -v ./...

setup:
	# install kube-api binary
	mkdir -p $(TESTBIN_DIR)
	rm -rf tmp && mkdir -p tmp
	curl -sfL https://github.com/kubernetes/kubernetes/archive/v$(K8S_VER).tar.gz | tar zxf - -C tmp
	mv tmp/kubernetes-$(K8S_VER) tmp/kubernetes
	cd tmp/kubernetes; make all WHAT="cmd/kube-apiserver"
	mv tmp/kubernetes/_output/bin/kube-apiserver $(TESTBIN_DIR)
	rm -rf tmp
	# install etcd binary
	curl -L https://storage.googleapis.com/etcd/v$(ETCD_VER)/etcd-v$(ETCD_VER)-linux-amd64.tar.gz | tar zx -C $(TESTBIN_DIR) --strip-components=1 etcd-v$(ETCD_VER)-linux-amd64/etcd
	# install kubectl binary
	curl -LO https://storage.googleapis.com/kubernetes-release/release/v$(K8S_VER)/bin/linux/amd64/kubectl
	chmod +x ./kubectl
	mv ./kubectl $(TESTBIN_DIR)/kubectl

$(TOPOLVM_CRD):
	mkdir -p $(CRD_DIR)
	curl -L -o $(TOPOLVM_CRD) https://raw.githubusercontent.com/topolvm/topolvm/v$(TOPOLVM_VER)/config/crd/bases/topolvm.cybozu.com_logicalvolumes.yaml
