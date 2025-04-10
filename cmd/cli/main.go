// Package main is a template for running a sensor controlled base without a viam-server
package main

import (
	"context"
	"controlledcomponents"

	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	goutils "go.viam.com/utils"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("cli")

	deps := resource.Dependencies{}
	// can load these from a remote machine if you need

	cfg := controlledcomponents.SCBConfig{}

	thing, err := controlledcomponents.NewSensorControlled(ctx, deps, base.Named("foo"), &cfg, logger)
	if err != nil {
		return err
	}
	defer func() { goutils.UncheckedError(thing.Close(ctx)) }()

	return nil
}
