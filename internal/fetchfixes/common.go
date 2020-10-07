package fetchfixes

import (
	"fmt"
	"os"
	"path"

	k8sserial "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// PersistObjectToYaml persists the given object to a yaml file in the given path
func persistObjectToYaml(name string, obj *unstructured.Unstructured, outputPath string, serializer *k8sserial.Serializer) (string, error) {
	path := path.Join(outputPath, name+".yaml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}

	defer f.Close()

	writer := json.YAMLFramer.NewFrameWriter(f)
	// Needed for MachineConfigs
	if obj.GetName() == "" {
		obj.SetName(name)
	}
	if err := serializer.Encode(obj, writer); err != nil {
		return "", fmt.Errorf("Couldn't serialize fix from rule '%s': %s", name, err)
	}

	if err = f.Sync(); err != nil {
		return "", err
	}
	return path, nil
}
