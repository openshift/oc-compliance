package fetchfixes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func setFixName(obj *unstructured.Unstructured, name string, id int, needsSuffix bool) {
	if !needsSuffix {
		obj.SetName(name)
	} else {
		obj.SetName(fmt.Sprintf("%s-%d", name, id))
	}
}

func PersistObject(name string, obj *unstructured.Unstructured, outputPath string) error {
	return nil
}
