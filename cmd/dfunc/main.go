// Package main
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  main.go
package main

import (
	"time"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/cmd/dfunc/controllers"
	"github.com/GabeCordo/DistributedFunctions/cmd/dfunc/local"
)

func main() {

	cli := commandline.NewCommandLine()

	bannerFactor := local.NewBannerFactory()
	cli.AddBanner(bannerFactor.Create(time.Now()))

	ic := cli.AddCommand("init", controllers.InitCommand{})
	ic.SetCategory("utils").SetDescription("initialize the global files required to start the DistributedFunctions")

	dc := cli.AddCommand("doctor", controllers.DoctorCommand{})
	dc.SetCategory("utils").SetDescription("verify the integrity of the global files on the log system")

	sc := cli.AddCommand("machine", controllers.StartCommand{})
	sc.SetCategory("utils").SetDescription("start the DistributedFunctions on the host")

	prc := cli.AddCommand("processors", controllers.ProcessorController{})
	prc.SetCategory("machine").SetDescription("list the processors connected to the core")

	mc := cli.AddCommand("modules", controllers.ModuleController{})
	mc.SetCategory("machine").SetDescription("list the modules mounted on the processors connected to the core")

	fc := cli.AddCommand("functions", controllers.FunctionController{})
	fc.SetCategory("machine").SetDescription("list the functions available on the processors connected to the core")

	runc := cli.AddCommand("run", controllers.RunController{})
	runc.SetCategory("machine").SetDescription("run a pipeline on the gateway")

	ppc := cli.AddCommand("pipelines", controllers.PipelinesController{})
	ppc.SetCategory("machine").SetDescription("bulk pipeline operations")

	pc := cli.AddCommand("pipeline", controllers.PipelineController{})
	pc.SetCategory("machine").SetDescription("single pipeline operations")

	schc := cli.AddCommand("schedule", controllers.ScheduleController{})
	schc.SetCategory("machine").SetDescription("schedule jobs on the gateway")

	psc := cli.AddCommand("ps", controllers.PSController{})
	psc.SetCategory("machine").SetDescription("view the running pipelines")

	mtc := cli.AddCommand("mount", controllers.MountController{})
	mtc.SetCategory("mounting").SetDescription("mount a module or function on the processors connected to the core")

	umtc := cli.AddCommand("unmount", controllers.UnMountController{})
	umtc.SetCategory("mounting").SetDescription("unmount a module or function from the processors connected to the core")

	cli.Run()
}
