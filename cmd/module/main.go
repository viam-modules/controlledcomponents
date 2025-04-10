// Package main runs modular main for the module
package main

import (
	"controlledcomponents"

	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{base.API, controlledcomponents.SensorControlledModel})
}
