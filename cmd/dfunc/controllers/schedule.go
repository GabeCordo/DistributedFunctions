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

const defaultFilePerm = 0600

type ScheduleController struct {
}

func (controller ScheduleController) Run(cli *commandline.CommandLine) (out commandline.TerminateOnCompletion) {

	if cli.Flag(commandline.Create) {
		out = controller.createScheduleJobs(cli)
	} else if cli.Flag(commandline.Delete) {
		out = controller.deleteScheduleJobs(cli)
	} else {
		out = controller.showScheduleJobs(cli)
	}

	return out
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
		fmt.Println("missing namespace identifier")
		return commandline.Terminate
	}

	namespace := cli.NextArg()
	if namespace == commandline.FinalArg {
		fmt.Println("missing namespace parameter")
		return commandline.Terminate
	}

	pipeline := cli.NextArg()
	if pipeline == commandline.FinalArg {
		fmt.Println("missing pipeline parameter")
		return commandline.Terminate
	}

	minute, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Println("missing or invalid minute parameter")
		return commandline.Terminate
	}

	hour, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Println("missing or invalid hour parameter")
		return commandline.Terminate
	}

	day, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Println("missing or invalid day parameter")
		return commandline.Terminate
	}

	month, ok := controller.parseIntervalValue(cli.NextArg())
	if !ok {
		fmt.Println("missing or invalid month parameter")
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

// parseIntervalValue
// converts one of the interval arguments into an int, rejecting values that
// do not fit in a uint8. A '*' argument matches the crontab wildcard and is
// stored as the core's wildcard value 0.
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
		fmt.Println("missing job identifier parameter")
		return commandline.Terminate
	}

	if err := api.DeleteJob(identifier); err != nil {
		fmt.Printf("Failed to delete the scheduled job '%s': %s\n", identifier, err)
		return commandline.Terminate
	}

	fmt.Printf("The job '%s' has been removed from the dfunc machine\n", identifier)

	return commandline.Terminate
}
