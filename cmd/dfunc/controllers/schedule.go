// Package controllers
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  schedule.go
package controllers

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/api"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
)

type ScheduleController struct {
}

func (controller ScheduleController) Run(cli *commandline.CommandLine) (action commandline.TerminateOnCompletion) {

	if cli.Flag(commandline.Create) {
		action = controller.createScheduleJobs(cli)
	} else if cli.Flag(commandline.Delete) {
		action = controller.deleteScheduleJobs(cli)
	} else {
		action = controller.showScheduleJobs(cli)
	}

	return action
}

func (controller ScheduleController) showScheduleJobs(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	jobs, err := api.GetJobs()
	if err != nil {
		fmt.Printf("Failed to retrieve the scheduled jobs from the dfunc machine: %s\n", err)
		return commandline.Terminate
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "JOB ID\tNAMESPACE\tPIPELINE\tMINUTE\tHOUR\tDAY\tMONTH")
	for _, j := range jobs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%d\n",
			j.Identifier, j.Namespace, j.Pipeline,
			j.Interval.Minute, j.Interval.Hour, j.Interval.Day, j.Interval.Month)
	}

	if err := w.Flush(); err != nil {
		fmt.Printf("Failed to print the scheduled jobs: %s\n", err)
	}

	return commandline.Terminate
}

func (controller ScheduleController) createScheduleJobs(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	identifier := cli.NextArg()
	if identifier == commandline.FinalArg {
		fmt.Println("Missing 'identifier' Parameter: dfunc schedule create [identifier] [namespace] [pipeline] [minute] [hour] [day] [month]")
		return commandline.Terminate
	}

	namespace := cli.NextArg()
	if namespace == commandline.FinalArg {
		fmt.Printf("Missing 'namespace' Parameter: dfunc schedule create %s [namespace] [pipeline] [minute] [hour] [day] [month]\n", identifier)
		return commandline.Terminate
	}

	pipeline := cli.NextArg()
	if pipeline == commandline.FinalArg {
		fmt.Printf("Missing 'pipeline' Parameter: dfunc schedule create %s %s [pipeline] [minute] [hour] [day] [month]\n", identifier, namespace)
		return commandline.Terminate
	}

	minute, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Printf("Missing or invalid 'minute' Parameter: dfunc schedule create %s %s %s [minute] [hour] [day] [month]\n",
			identifier, namespace, pipeline)
		return commandline.Terminate
	}

	hour, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Printf("Missing or invalid 'hour' Parameter: dfunc schedule create %s %s %s %d [hour] [day] [month]\n",
			identifier, namespace, pipeline, minute)
		return commandline.Terminate
	}

	day, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Printf("Missing or invalid 'day' Parameter: dfunc schedule create %s %s %s %d %d [day] [month]\n",
			identifier, namespace, pipeline, minute, hour)
		return commandline.Terminate
	}

	month, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Printf("Missing or invalid 'month' Parameter: dfunc schedule create %s %s %s %d %d %d [month]\n",
			identifier, namespace, pipeline, minute, hour, day)
		return commandline.Terminate
	}

	jb := job.Job{
		Identifier: identifier,
		Namespace:  namespace,
		Pipeline:   pipeline,
		Interval: database.Interval{
			Minute: minute,
			Hour:   hour,
			Day:    day,
			Month:  month,
		},
	}

	if err := api.CreateJob(jb); err != nil {
		fmt.Printf("Failed to create the scheduled job: %s\n", err)
		return commandline.Terminate
	}

	fmt.Printf("The job '%s' has been scheduled on the dfunc machine\n", jb.Identifier)

	return commandline.Terminate
}

func (controller ScheduleController) parseIntervalValue(input string) (int, bool) {

	if input == "*" {
		return 0, true
	}

	value, err := strconv.ParseUint(input, 10, 8)
	if err != nil {
		fmt.Println(err)
		return 0, false
	}

	return int(value), true
}

func (controller ScheduleController) deleteScheduleJobs(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	identifier := cli.NextArg()
	if identifier == commandline.FinalArg {
		fmt.Println("Missing 'identifier' Parameter: dfunc schedule delete [identifier]")
		return commandline.Terminate
	}

	if err := api.DeleteJob(identifier); err != nil {
		fmt.Printf("Failed to delete the scheduled job '%s': %s\n", identifier, err)
		return commandline.Terminate
	}

	fmt.Printf("The job '%s' has been removed from the dfunc machine\n", identifier)

	return commandline.Terminate
}
