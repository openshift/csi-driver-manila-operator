package controller

import (
	"github.com/openshift/csi-driver-manila-operator/pkg/controller/manilacsi"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, manilacsi.Add)
}
