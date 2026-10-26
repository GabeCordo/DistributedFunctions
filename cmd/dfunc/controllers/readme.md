# controllers

Command Line Interface (CLI) controllers encapsulate functionality associated
with a unique identifier. For example, if the processes takes in the parameter 'start',
a start controller is defined such that anytime the identifier is invoked by an operator, the funciton
will be called.

Controllers are grouped by the category they are registered under in `main.go`.

## utils

### init
Create the required temporary files and general config used by DistributedFunctions.

### doctor
Verify the required temporary files have been created and the general config used by DistributedFunctions is valid.

### machine
Manage the DistributedFunctions process on the host.
- `dfunc machine start`: launch the dfunc machine in the background
- `dfunc machine start --blocking`: launch the dfunc machine in the foreground
- `dfunc machine stop`: terminate a running dfunc machine
- `dfunc machine status`: display whether the dfunc machine is online and its PID

## machine

### processors
List the processors connected to the dfunc machine.

### modules
List the modules registered by processors on the dfunc machine.

### functions
List the functions belonging to modules registered by processors on the dfunc machine.

### run
Run pipelines on a processor connected to the dfunc machine.
- `dfunc run [namespace] [pipeline]`: run a pipeline; the identifier of the created run is printed so it can be used by automated tooling
- `dfunc stop run [id]`: terminate a run by its identifier

### pipelines
List pipelines that are made of one-to-many functions that are runnable on processors.

### schedule
List pipelines that are run on a scheduled interval.
- `dfunc schedule`: list the scheduled jobs
- `dfunc schedule create [identifier] [namespace] [pipeline] [minute] [hour] [day] [month]`: create a scheduled job; a `*` matches any value for an interval argument
- `dfunc schedule delete [identifier]`: remove a scheduled job

### ps
View pipelines that are actively running on processors or have completed.
- `dfunc ps`: list only the running pipelines
- `dfunc ps -a`: list all pipelines, including those that have finished

## mounting

### mount
Mount a module or function on the processors connected to the core. Mounting allows a module or function to be run.
- `dfunc mount module [module]`
- `dfunc mount function [module] [function]`

### unmount
Unmount a module or function from the processors connected to the core. Unmounting prevents a module or function from being run.
- `dfunc unmount module [module]`
- `dfunc unmount function [module] [function]`
