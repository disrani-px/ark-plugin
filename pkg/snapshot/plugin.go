package snapshot

import (
	"fmt"

	"github.com/heptio/ark/pkg/cloudprovider"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/portworx/sched-ops/k8s"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// serviceName is the name of the portworx service
	serviceName = "portworx-service"

	// namespace is the kubernetes namespace in which portworx
	// daemon set
	// runs
	namespace = "kube-system"

	// Config parameters
	configType = "Type"
	configCred = "CredId"

	typeLocal = "Local"
	typeCloud = "Cloud"
)

func getVolumeDriver() (volume.VolumeDriver, error) {
	var endpoint string
	svc, err := k8s.Instance().GetService(serviceName, namespace)
	if err == nil {
		endpoint = svc.Spec.ClusterIP
	} else {
		return nil, fmt.Errorf("Failed to get k8s service spec: %v", err)
	}

	if len(endpoint) == 0 {
		return nil, fmt.Errorf("Failed to get endpoint for portworx volume driver")
	}

	clnt, err := volumeclient.NewDriverClient("http://"+endpoint+":9001", "pxd", "", "stork")
	if err != nil {
		return nil, err
	}
	return volumeclient.VolumeDriver(clnt), nil
}

type SnapshotPlugin struct {
	Log    logrus.FieldLogger
	plugin cloudprovider.BlockStore
}

func (s *SnapshotPlugin) Init(config map[string]string) error {
	s.Log.Infof("Init'ing portworx plugin with config %v", config)
	if snapType, ok := config[configType]; !ok || snapType == typeLocal {
		s.plugin = &localSnapshotPlugin{log: s.Log}
	} else {
		err := fmt.Errorf("Snapshot type %v not supported", snapType)
		s.Log.Errorf("%v", err)
		return err
	}

	return s.plugin.Init(config)
}

func (s *SnapshotPlugin) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	return s.plugin.CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ, iops)
}

func (s *SnapshotPlugin) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	return s.plugin.GetVolumeInfo(volumeID, volumeAZ)
}

func (s *SnapshotPlugin) IsVolumeReady(volumeID, volumeAZ string) (ready bool, err error) {
	return s.plugin.IsVolumeReady(volumeID, volumeAZ)
}

func (s *SnapshotPlugin) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	return s.plugin.CreateSnapshot(volumeID, volumeAZ, tags)
}

func (s *SnapshotPlugin) DeleteSnapshot(snapshotID string) error {
	return s.plugin.DeleteSnapshot(snapshotID)
}

func (s *SnapshotPlugin) GetVolumeID(pv runtime.Unstructured) (string, error) {
	return s.plugin.GetVolumeID(pv)
}

func (s *SnapshotPlugin) SetVolumeID(pv runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	return s.plugin.SetVolumeID(pv, volumeID)
}
