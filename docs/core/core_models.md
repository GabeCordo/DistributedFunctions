# Core Models
The data models are stored inside the core using `in-memory`, `mongodb` or `sqlite` datastores. 

## Contacts

A `contact` represents a subscriber for event notifications.

```mermaid
erDiagram
    CONTACT {
        string identifier PK
        string emails
        string subscriptions
    }
```

## Jobs
A `job` represents a scheduable task the core is responsible for triggering.

```mermaid
erDiagram
    JOB {
        string id PK
        string identifier
        string namespace
        string pipeline
    }
    JOB_INTERVAL {
        string id PK
        string job_id FK
        int minute
        int hour
        int day
        int month
    }
    JOB_METADATA {
        string id PK
        string job_id FK
        string key
        string value
    }
    JOB ||--o{ JOB_INTERVAL : contains
    JOB ||--o{ JOB_METADATA : contains
```

## Pipelines

A `pipeline` represents a set of functions and pipes that define a data processing workflow.

```mermaid
erDiagram
    PIPELINE {
        string id PK
        string namespace
        string identifier
    }
    PIPELINE_IR {
        string id PK
        string pipeline_id FK
        string on_crash
        string on_startup
        string on_teardown
    }
    FUNCTION_IR {
        string id PK
        string pipeline_ir_id FK
        string module
        string identifier
        string from_pipe
        string to_pipe
        int start_with
        bool wait_before
        int maximum
        string parameters
        string returns
        bool static_mount
    }
    PIPE_IR {
        string id PK
        string pipeline_ir_id FK
        string identifier
        int threshold
        float growth_factor
    }
    PIPELINE ||--o{ PIPELINE_IR : contains
    PIPELINE_IR ||--o{ FUNCTION_IR : contains
    PIPELINE_IR ||--o{ PIPE_IR : contains
```

## Runs

A `run` represents an execution of a pipeline.

```mermaid
erDiagram
    RUN {
        string id PK
        string status
        string started_by
        uint64 processor
        string namespace
        string pipeline
        string statistics
        datetime created
        datetime last_updated
        uint64 duration_hours
        uint64 duration_minutes
        uint64 duration_seconds
        uint64 duration_milliseconds
    }
```

## Statistics

A `statistic` represents runtime statistics for a pipeline execution.

```mermaid
erDiagram
    STATISTIC {
        string id PK
        string timestamp
        string namespace
        string pipeline
        int elapsed_ns
        int num_of_functions
        int num_of_channels
    }
    STATISTICS_DATA {
        string id PK
        string statistic_id FK
    }
    FUNCTION_STATISTIC {
        string id PK
        string statistics_data_id FK
        int index
        int active
        int provisions
    }
    PIPE_STATISTIC {
        string id PK
        string statistics_data_id FK
        int index
        int pushed
        int pulled
        int dropped
        int breaches
    }
    TIMING_STATISTIC {
        string id PK
        string pipe_statistic_id FK
        int min_time_before_pop_ns
        int max_time_before_pop_ns
        int average_time_ns
        int median_time_ns
    }
    STATISTIC ||--o{ STATISTICS_DATA : contains
    STATISTICS_DATA ||--o{ FUNCTION_STATISTIC : contains
    STATISTICS_DATA ||--o{ PIPE_STATISTIC : contains
    PIPE_STATISTIC ||--o{ TIMING_STATISTIC : contains
```
