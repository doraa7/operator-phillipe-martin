/*
Copyright 2018 Anevia.
*/

// NOTE: Boilerplate only.  Ignore this file.

// Package v1 contains API Schema definitions for the cluster v1 API group
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/feloy/operator/pkg/apis/cluster
// +k8s:defaulter-gen=TypeMeta
// +groupName=cluster.anevia.com
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "cluster.anevia.com", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
