# core-processor tunnel
This specification covers the TCP tunnel between the **core** and **processor** services.

You can implement the tunnel protocol to create custom **processor** services and communicate with the core.

## Tunnel Connection

The tunnel is a TCP connection established between a **core** and **processor**. The TCP connection uses JSON encoding to send messages between the endpoints.

## Tunnel Message

The JSON encoding contains three structures: (1) action, (2) record, and (3) data.

#### Message Action (uint8)
The **action** field defines what the initiating endpoint is intending to do.

1. Ping (0)
2. Create (1)
3. Update (2)
4. Delete (3)

#### Message Type (uint8)
The **record** field defines what the initiating endpoint is performing an action on.

1. Module (0)
2. Run (1)
3. Log (2)

#### Message Data (structure)
The **data** field defines information that the **action**/**record** pair are operating on.

##### Basic Example

```json
{
    "action": 0,
    "record": 0,
    "data": 0
}
```


## Tunnel Procedures
Events sent across the tunnel.

### Processor Connect
The **processor** initiates a TCP connection with the **core**. Establishing a TCP connection the **core** registers
it to a set of **processor** that may be responsible for performing compute. The **processor** after establishing
connection is assigned an empty set of modules implying that it cannot perform any compute.

### Processor Disconnect
The **processor** disconnects the TCP connection with the **core** to remove itself from the set of **processors**
registered to the core.

### Processor Create Module
The **processor** sends a JSON encoded payload to the **core** with the `action:create` and `type:module` parameters
to add a module to its set of runnable modules.

#### Message JSON Payload

```json
{
    "action": 1,
    "record": 0,
    "data": {
        "id": "",
        "version": "",
        "contact": {
            "name": "",
            "email": ""
        },
        "functions": [
            {
                "module": "",
                "id": "",
                "metadata": {
                    "static_mount": false
                },
                "from": "",
                "to": "",
                "start_with": "",
                "wait_before": "",
                "maximum": "",
                "parameters": [ "" ],
                "returns": [ "" ]
            }
        ]
    } 
}
```

### Core Create Run
The **core** sends a JSON encoded payload to the **processor** with the `action:create` and `type:run` parameters to run a function that belongs to its set of registers modules.

#### Message JSON Payload
```json
{
    "action": 1,
    "record": 1,
    "data": {
        "id": 0,
        "namespace": "",
        "config": {
            "id": "",
            "on_crash": "",
            "functions": [
                {
                    "module": "",
                    "id": "",
                    "metadata": {
                        "static_mount": false
                    },
                    "from": "",
                    "to": "",
                    "start_with": "",
                    "maximum": "",
                    "parameters": [ "" ],
                    "returns": [ "" ]
                }
            ],
            "pipes": [
                {
                    "id": "",
                    "threshold": "",
                    "growth_factor": ""
                }
            ],
            "on_startup": "",
            "on_teardown": ""
        },
        "metadata": { 
            "": "" 
        }
    }
}
```

### Processor Update Run
The **processor** sends a JSON encoded payload to the **core** with the `action:update` and `type:run` parameters
to inform the **core** the state of a run it created.

#### Message JSON Payload
```json
{
    "action": 2,
    "record": 1,
    "data": {
        "id": "",
        "namespace": "",
        "processor": "",
        "status": "",
        "time": {
            "created": "",
            "last_updated": "",
            "duration": {
                "hours": 0,
                "minutes": 0,
                "seconds": 0,
                "milliseconds": 0
            }
        },
        "started_by": "",
        "statistics": {
            "num_of_functions": 0,
            "functions": [
                {
                    "active": 0,
                    "provisions": 0
                }
            ],
            "num_of_channels": 0,
            "channels": [
                {
                    "processed": 0,
                    "dropped": 0,
                    "breaches": 0,
                    "timing": {
                        "min_time_before_pop_ns": 0,
                        "max_time_before_pop_ns": 0,
                        "average_time_ns": 0,
                        "median_time_ns": 0
                    }
                }
            ]
        }
    }
}
```

### Core Stops Run
The **core** sends a JSON encoded payload to the **processor** with the `action:delete` and `type:run` parameters to stop an
exisiting run that was created on a processor.

#### Message JSON Payload
```json
{
    "action": 3,
    "record": 1,
    "data": 0
}
```

## Data Types

The data types sent between the processor and core on the TCP connection.


###### Module
The module structure carries information about the runnable functions available on the processor.

```text
module => {
    id : string
    version : string
    contact => {
        name : string
        email : string
    }
    functions => []{
        id : string
        parameters => []string
        returns => []string
    }
}
```

###### Run
The run structure carries information about the running function(s) on the prcessor.

```text
run => {
    id : uint64
    status => enum{
        created
        activated
        crashed
        completed
        terminated
        cancelled
    }
    time => {
        created : time (optional)
        last_updated : time (optional)
        duration => {
            hours : uint64
            minutes : uint64
            seconds : uint64
            milliseconds : uint64
        } (optional)
        started_by : => enum{
            operator
            processor
            scheduler
        }
        processor : uint64 (optional)
        namespace : string (optional)
        pipeline => {
            identifier : string
            on_crash => enum {
                Restart
                DoNothing
            }
            functions => []{
                module : string
                identifier : string
                metadata => {
                    static_mount : bool
                }
                from : string (optional)
                to : string (optional)
                start_with : uint16 (optional)
                wait_before : bool (optional)
                maximum : uint16 (optional)
                parameters => []string
                returns => []string
            }
            pipes => []{
                identifier : string
                threshold : uint32
                growth_factor : float64
            }
        } (optional)
        statistics => {
            num_of_functions : uint16
            functions => []{
                active : uint16
                provisions : uint16
            }
            num_of_pipes : uint16
            pipes => []{
                pushed : uint64
                pulled : uint64
                dropped : uint64
                breaches : uint64
                timing => {
                    min_time_before_pop_ns : uint64
                    max_time_before_pop_ns : uint64
                    average_time_ns : uint64
                    median_time_ns : uint64
                }
            }
        } (optional)
    }
}
```

###### Log
Log has not been designed.

### Message Format
The message sent between the services is formated as such. The _type_ field specifies how the service should decode the 
content of the _data_ field.

```text
message => {
    action : MessageAction
    type : MessageType
    data : Module | Run | Log
}
```