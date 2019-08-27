package controller

import (
	"github.com/linki/cloudformation-operator/pkg/controller/stack"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, stack.Add)
}
