// Package main
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  main.go
package main

import (
	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/cmd/dfunc/controllers"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
)

var banner = "    ___ _     _        _ _           _           _   ___                 _   _  \n" +
	"   /   (_)___| |_ _ __(_) |__  _   _| |_ ___  __| | / __\\   _ _ __   ___| |_(_) ___  _ __  ___  \n" +
	"  / /\\ / / __| __| '__| | '_ \\| | | | __/ _ \\/ _` |/ _\\| | | | '_ \\ / __| __| |/ _ \\| '_ \\/ __|\n" +
	" / /_//| \\__ \\ |_| |  | | |_) | |_| | ||  __/ (_| / /  | |_| | | | | (__| |_| | (_) | | | \\__ \\\n" +
	"/___,' |_|___/\\__|_|  |_|_.__/ \\__,_|\\__\\___|\\__,_\\/    \\__,_|_| |_|\\___|\\__|_|\\___/|_| |_|___/\n" +
	"[+] " + terminal.Purple + "Distributed Functions " + terminal.Reset + "\n" +
	"[+]" + terminal.Purple + " by Gabriel Cordovado 2022-2026" + terminal.Reset + "\n"

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
func main() {

	cli := commandline.NewCommandLine()
	cli.AddBanner(banner)

	ic := cli.AddCommand("init", controllers.InitCommand{})
	ic.SetCategory("utils").SetDescription("initialize the global files required to start the DistributedFunctions")

	dc := cli.AddCommand("doctor", controllers.DoctorCommand{})
	dc.SetCategory("utils").SetDescription("verify the integrity of the global files on the log system")

	lc := cli.AddCommand("logs", controllers.LogController{})
	lc.SetCategory("data").SetDescription(
		"used to view log files stored on the log system" +
			"\n\t\t[path] specify the name of the log file" +
			"\n\t\t[normal|warning|fatal] specify the log priority " +
			"\n\t\t\t(ex. find fatal errors)")

	stc := cli.AddCommand("stats", controllers.StatisticsController{})
	stc.SetCategory("data").SetDescription("used to view the statistic files stored on the system")

	sc := cli.AddCommand("start", controllers.StartCommand{})
	sc.SetCategory("utils").SetDescription("start the DistributedFunctions on the log system")

	rc := cli.AddCommand("repl", controllers.ReplController{})
	rc.SetCategory("utils").SetDescription("enable or disable the repl when running mango start")

	shc := cli.AddCommand("scheduler", controllers.ScheduleController{})
	shc.SetCategory("utils").SetDescription("create or delete schedules for when clusters should be provisioned")

	cc := cli.AddCommand("config", controllers.ConfigCommand{})
	cc.SetCategory("utils").SetDescription("update the global DistributedFunctions configuration")

	// util controllers

	ec := cli.AddCommand("example", controllers.ExampleController{})
	ec.SetCategory("utils").SetDescription("an example processor node to test the gateway")

	runc := cli.AddCommand("run", controllers.RunController{})
	runc.SetCategory("utils").SetDescription("run a pipeline on the gateway")

	pc := cli.AddCommand("pipeline", controllers.PipelineController{})
	pc.SetCategory("utils").SetDescription("register a pipeline on the gateway")

	schc := cli.AddCommand("schedule", controllers.ScheduleController{})
	schc.SetCategory("utils").SetDescription("schedule jobs on the gateway")

	statc := cli.AddCommand("jobs", controllers.StatsController{})
	statc.SetCategory("utils").SetDescription("view stats associated with jobs on DistributedFunctions")

	// configuration controllers

	nc := cli.AddCommand("namespace", controllers.NamespaceController{})
	nc.SetCategory("config").SetDescription("switch the namespace used by pops")

	gc := cli.AddCommand("gate", controllers.GatewayController{})
	gc.SetCategory("config").SetDescription("switch the gateway used by pops")

	cli.Run()
}
