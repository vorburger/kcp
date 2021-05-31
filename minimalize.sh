rm -rf cmd/cluster-controller
rm -rf cmd/crd-puller
rm -rf cmd/deployment-splitter
rm -rf cmd/kcp/
rm -rf cmd/syncer

rm -rf pkg/apis/
rm -rf pkg/client/
rm -rf pkg/crdpuller/
rm -rf pkg/reconciler/
rm -rf pkg/syncer/
rm -rf pkg/util/

rm -f cluster.yaml
rm -f ns.yaml

# TODO what is still required in go.mod after this minimalization, for kcp-core to still work?
