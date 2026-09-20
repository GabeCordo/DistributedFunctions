# core
The core is a centralized service for receiving, distributing, and monitoring requests to run data pipelines.

The core was designed to simplify the deployment and monitoring of data pipelines.


---

## Users
The users of the core are the [operator](#operator-user) and [developer](#developer-user). It's important to
note that these two users may be the same person depending on the size of the team.

### Operator User
The operator is responsible for the production deployment and monitoring of data pipelines.

The operator will be abbreviated as "O" within the document.

### Developer User
The developer is responsible for building and iterating on data pipelines.

The developer will be abbreviated as "D" within the document.

### Product Owner User
The product owner is responsible for the completion of features and billing. 

The product owner will be abbreviated as "PO" within the document.

---


## Goals
Goals are prefixed with "G" followed by the abbreviation of the user and unique identifier per user.

| Identifier | Status   | Achieved In Version | Link |
| :- |:---------|:--------------------|:-----|
| G01 | achieved | todo                |  |
| G02 | missing |                  |      |
| G03 | achieved | todo | |
| G04 | missing |  | |
| GD1 | achieved | | | |
| GD2 | achieved | | | |
| GD3 | missing | | | |
| GP1 | missing | | | |
| GP2 | missing | | | |
| GP3 | missing | | | |
_Table. The goals of the core process._

### Operator Goals
The goals of the [operator](#operator-user) user.

###### GO1
Make a pipeline available to external users and applications.

###### GO2
Know when pipelines have an uptick in faults.

###### GO3
Disable pipelines that show an uptick in faults.

###### GO4
Rollback pipelines to a previous working version when the latest has issues.

### Developer Goals
The goals of the [developer](#developer-user) user.

###### GD1
Develop data pipelines that process .

###### GD2
Monitor data flow through the pipeline to find bottlenecks.

###### GD3
View logs generated as data flows through the pipeline.

###### GD4
Have the resources of the data pipeline automatically scale to meet the demand of the data flowing through the pipeline.

### Product Owner Goals
The goals of the [product owner](#product-owner-user) user.

###### GP1
Assign key performance metrics to a data pipeline. 

###### GP2
View key performance metrics to a data pipeline.

###### GP3
Know when key performance metrics for a data pipeline have been breached.

---


## Requirements
Requirements are prefixed with "R" followed by a unique identifier per requirement.

| Identifier | Achieved | Achieved in Version |
| :- |:---------|:--------------------|
| R0 | Th

###### R0
The core shall represent executable code as a function.

###### R1
The core shall store the parameter types of a function.

###### R2
The core shall store the types returned from a function.

###### R3
The core shall represent a set of functions as a module.

###### R4
The core shall version each module to track incremental changes.

###### R5
The core shall define code that can be run as mounted. 

###### R6
The core shall define code that cannot be run as unmounted.

###### R7
The core shall allow an operator to mount a module.

###### R8
The core shall allow an operator to unmount a module.

###### R9
The core shall allow an operator to mount a function.

###### R10
The core shall allow an operator to unmount a function.

###### R11
The core shall define a set of functions in a pipeline.

###### R12
The core shall define a set of pipes in a pipeline.

###### R13
The core may define a pipe a function receives data from in a pipeline.

###### R14
The core may define a pipe that a function sends data to in a pipeline.

###### R15
The core shall define a pipeline as the smallest executable unit.

> When send a request for code to run on the core, we are sending a request to run a pipeline rather than a singular or set of functions. 

###### R16
The core shall verify each function in a pipeline is mounted before code is executed.

###### R17
The core shall find a processor that supports each function in a pipeline so code may be executed.

###### R18
The core shall send a request to a [processor](#processor) to run a [pipeline](#pipeline). 

---

## Models
These models are standardized representations to store data inside the core.

### Processor
A processor is an external computer process that connects to the core to provide compute resources.

The processor exposes [modules](#module) that contain a set of runnable [functions](#function) on the core.

###### Processor Status
```json
[
  "active"
]
```

###### Processor Record
```json
{
  "Id": "string",
  "RemoteAddr": "string",
  "Status": "ProcessorStatus",
  "LastUpdate": "time",
  "Modules": "list[str]",
  "Retries": 0,
  "NumOfRuns": 0
}
```

### Module

### Function

These modules and functions can be [mounted](#mount-function) or [unmounted](#unmount-function) to
control whether they can be run on a processor.


### Pipeline

### Run

### Statistic