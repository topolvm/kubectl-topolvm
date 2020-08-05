K8S_VER=1.18.6
ETCD_VER=3.3.22
TESTBIN_DIR=bin

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
	curl -L https://github.com/etcd-io/etcd/releases/download/v$(ETCD_VER)/etcd-v$(ETCD_VER)-linux-amd64.tar.gz | tar zx -C $(TESTBIN_DIR) --strip-components=1 etcd-$(ETCD_VER)-linux-amd64/etcd
	# install kubectl binary
	curl -LO https://storage.googleapis.com/kubernetes-release/release/v$(K8S_VER)/bin/linux/amd64/kubectl
	chmod +x ./kubectl
	mv ./kubectl $(TESTBIN_DIR)/kubectl
