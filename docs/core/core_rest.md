## Core REST APIs
Defines the endpoints used to query information inside the core.

### /processor

#### GET
Fetch a set of [processor](#11-processor) records from the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/processor
```

---


### /module

#### GET
Fetch a set of [module](#12-module) records from the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/module
```

#### PUT

##### Mount Module
Allow an operator to run all functions inside a module. 

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/module -H "Content-Type: application/json" -d "{\"module\":\"common\",\"mounted\":true}"
```

##### Unmount Module
Disable an operator from running all functions inside a module. 

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/module -H "Content-Type: application/json" -d "{\"module\":\"common\",\"mounted\":false}"
```

---


### /function
Fetch a set of [function](#13-function) records from the core.

#### GET

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/function\?module=common
```

#### PUT

##### Mount Function
Allow an operator to run a function inside a module.

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/function -H "Content-Type: application/json" -d "{\"module\":\"common\",\"function\":\"prt\",\"mounted\":true}"
```

##### Unmount Function
Disable an operator from running a function inside a module.

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/function -H "Content-Type: application/json" -d "{\"module\":\"common\",\"function\":\"prt\",\"mounted\":false}"
```

---


### /namespaces

#### GET
Fetch a set of namespaces used in the core. Namespaces are used to allow scoped use of pipeline and job identifiers rather
than maintaining a global namespace.

The _common_ namespace is the default global namespace that is present in all cores.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/namespaces
```


---


### /pipeline

#### GET

###### HTTP Params
namespace 

pipeline _(optional)_

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/pipeline\?namespace=common
```

#### POST
Create a new pipeline that defines the flow of data between functions run on processors.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/pipeline -H "Content-Type: application/json" -d @docs/core/examples/pipelines/hello-world.json
```

#### PUT
Update a pipeline stored on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X PUT -L http://127.0.0.1:8136/pipeline -H "Content-Type: application/json" -d @docs/core/examples/pipelines/hello-world.json
```

#### DELETE
Delete a pipeline stored on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X DELETE -L http://127.0.0.1:8136/pipeline\?namespace=common\&pipeline=hello-world
```


---


### /run

#### GET
Fetch the data received from a processor for a running pipeline.

###### HTTP Params
namespace (mandatory)

pipeline (optional)

maximumResults (optional)

offsetOfResults (optional)

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run\?namespace=common\&pipeline=hello-world
```

#### POST
Provision a new running instance of a pipeline.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/run -H "Content-Type: application/json" -d @docs/core/examples/runs/hello-world.json
```

### /run/count

#### GET
Get the number of runs for a pipeline.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run/count\?namespace=common\&pipeline=hello-world
```

---


### /run/statistic

#### GET
Fetch the statistics collected from completed pipeline runs.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run/statistic\?namespace=common\&pipeline=hello-world
```


---


### /run/statistic/info

#### GET

##### Retrieve Namespaces for Statistics

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run/statistic/info
```

##### Retrieve Pipeline Statistics for a Namespace

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/run/statistic/info\?namespace=common
```


---


### /job
Jobs define a schedule where the core will provision runs for a pipeline.  

#### GET
Fetch the jobs present on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/job\?namespace=common
```

#### POST
Create a job on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/job -H "Content-Type: application/json" -d @docs/core/examples/jobs/hello-world.json
```

#### DELETE
Delete a job on the core.

###### CURL Example
```bash
curl -H "Accept: application/json" -X DELETE -L http://127.0.0.1:8136/job\?id=hello-job
```


---


### /debug

#### GET
The endpoint shall be used to validate whether the core is online.

###### CURL Example
```bash
curl -H "Accept: application/json" -X GET -L http://127.0.0.1:8136/debug
```

#### POST

##### Shutdown

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/debug -H "Content-Type: application/json" -d "{\"action\":\"shutdown\"}"
```

##### Latency

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/debug -H "Content-Type: application/json" -d "{\"action\":\"ping\"}"
```

##### Toggle Debug

###### CURL Example
```bash
curl -H "Accept: application/json" -X POST -L http://127.0.0.1:8136/debug -H "Content-Type: application/json" -d "{\"action\":\"debug\"}"
```




