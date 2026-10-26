// Package api
//
// Copyright (c) 2026 Gabriel Cordovado
// All rights reserved
//
// Source file:  core.go
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/processor"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/run"
	"github.com/GabeCordo/ScalingFunctions"
)

func Connect() (err error) {

	err = nil

	url := fmt.Sprintf("http://localhost:5023/gate?=%saction&=connect", "default")

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", rsp.Status)
	}

	return err
}

func Disconnect() (err error) {

	err = nil

	url := fmt.Sprintf("http://localhost:5023/gate?=%saction&=disconnect", "default")
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", rsp.Status)
	}

	return err
}

func Ping() (pid int, err error) {

	pid = -1
	err = nil

	rsp, err := http.Get(fmt.Sprintf("%s/debug", coreHost))
	if err != nil {
		return pid, err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", rsp.Status)
		return pid, err
	}

	r := Response{}
	err = json.NewDecoder(rsp.Body).Decode(&r)
	if err != nil {
		return pid, err
	}

	if !r.Success {
		err = fmt.Errorf("the core did not respond to the ping")
		return pid, err
	}

	data, ok := r.Data.(map[string]any)
	if !ok {
		err = fmt.Errorf("the ping response is missing data")
		return pid, err
	}

	fPid, ok := data["pid"].(float64)
	if !ok {
		err = fmt.Errorf("the ping response does not contain a pid")
		return pid, err
	}

	pid = int(fPid)
	return pid, err
}

func GetProcessors() (processors []processor.Processor, err error) {

	processors = nil
	err = nil

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/processor", coreHost), nil)
	if err != nil {
		return processors, err
	}
	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return processors, err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return processors, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := dataResponse[[]processor.Processor]{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return processors, err
	}

	processors = response.Data
	return processors, err
}

func GetModules() (modules []processor.ModuleData, err error) {

	modules = nil
	err = nil

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/module", coreHost), nil)
	if err != nil {
		return modules, err
	}
	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return modules, err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return modules, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := dataResponse[[]processor.ModuleData]{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return modules, err
	}

	modules = response.Data
	return modules, err
}

type Function struct {
	Module     string
	Name       string
	Mounted    bool
	Parameters []string
	Returns    []string
}

func GetFunctions() (functions []Function, err error) {

	functions = nil
	err = nil

	modules, err := GetModules()
	if err != nil {
		return functions, err
	}

	functions = make([]Function, 0)

	for _, module := range modules {
		moduleFunctions, err := GetFunctionsInModule(module.Name)
		if err != nil {
			return functions, err
		}
		functions = append(functions, moduleFunctions...)
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	return functions, err
}

func GetFunctionsInModule(module string) (functions []Function, err error) {

	functions = nil
	err = nil

	url := fmt.Sprintf("%s/function?module=%s", coreHost, module)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return functions, err
	}
	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return functions, err
	}
	defer closeBody(rsp.Body)

	response := dataResponse[[]processor.FunctionData]{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return functions, err
	}

	functions = make([]Function, 0, len(response.Data))
	for _, f := range response.Data {
		functions = append(functions, Function{
			Module:     module,
			Name:       f.Name,
			Mounted:    f.Mounted,
			Parameters: f.Parameters,
			Returns:    f.Returns,
		})
	}

	return functions, err
}

func MountModule(module string) error {
	return setModuleMounted(module, true)
}

func UnmountModule(module string) error {
	return setModuleMounted(module, false)
}

func setModuleMounted(module string, mounted bool) (err error) {

	err = nil

	b, err := json.Marshal(struct {
		Module  string `json:"module"`
		Mounted bool   `json:"mounted"`
	}{module, mounted})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/module", coreHost), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", rsp.Status)
	}

	return err
}

func MountFunction(module, function string) error {
	return setFunctionMounted(module, function, true)
}

func UnmountFunction(module, function string) error {
	return setFunctionMounted(module, function, false)
}

func setFunctionMounted(module, function string, mounted bool) (err error) {

	err = nil

	b, err := json.Marshal(struct {
		Module   string `json:"module"`
		Function string `json:"function"`
		Mounted  bool   `json:"mounted"`
	}{module, function, mounted})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/function", coreHost), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return err
}

func IsPipelineOnCore(host, namespace, pl string) (found bool, err error) {

	found = false
	err = nil

	url := fmt.Sprintf("%s/pipeline?namespace=%s", host, namespace)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return found, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return found, err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode == http.StatusNotFound {
		return found, err
	}

	pipelines := make([]ScalingFunctions.PipelineIR, 0)
	err = json.NewDecoder(rsp.Body).Decode(&pipelines)
	if err != nil {
		return found, err
	}

	for _, p := range pipelines {
		if p.Identifier == pl {
			found = true
			return found, err
		}
	}

	return found, err
}

func CreatePipelineOnCore(host, namespace string, pl *ScalingFunctions.PipelineIR) (err error) {

	err = nil

	url := fmt.Sprintf("%s/pipeline?namespace=%s", host, namespace)

	b, err := json.Marshal(pl)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	}

	return err
}

func RemovePipelineFromCore(host, namespace, pipeline string) (err error) {

	err = nil

	url := fmt.Sprintf("%s/pipeline?namespace=%s&pipeline=%s", host, namespace, pipeline)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	}

	return err
}

func RunPipelineOnCore(host, namespace, pipeline string) (id uint64, err error) {

	id = 0
	err = nil

	url := fmt.Sprintf("%s/run", host)

	body := &struct {
		Namespace string            `json:"namespace"`
		Pipeline  string            `json:"pipeline"`
		Metadata  map[string]string `json:"metadata"`
	}{namespace, pipeline, make(map[string]string)}

	b, err := json.Marshal(body)
	if err != nil {
		return id, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return id, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return id, err
	}
	defer closeBody(rsp.Body)

	response := &runResponse{}

	if rsp.StatusCode != http.StatusOK {
		return id, errors.New("http status code " + rsp.Status)
	}

	err = json.NewDecoder(rsp.Body).Decode(response)
	if err != nil {
		return id, err
	}

	id = response.Id
	return id, err
}

func StopRun(host string, id uint64) (err error) {

	err = nil

	url := fmt.Sprintf("%s/run?id=%d", host, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return errors.New("something went wrong while stopping the run")
	}

	return nil
}

func GetNamespaces() (namespaces []string, err error) {

	namespaces = nil
	err = nil

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/namespaces", coreHost), nil)
	if err != nil {
		return namespaces, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return namespaces, err
	}

	response := dataResponse[[]string]{}
	decodeErr := json.NewDecoder(rsp.Body).Decode(&response)
	closeErr := rsp.Body.Close()
	if decodeErr != nil {
		return namespaces, decodeErr
	}
	if closeErr != nil {
		fmt.Print(closeErr)
	}

	namespaces = response.Data
	return namespaces, err
}

type Pipeline struct {
	Namespace  string
	Identifier string
}

func GetPipelines() (pipelines []Pipeline, err error) {

	pipelines = nil
	err = nil

	namespaces, err := GetNamespaces()
	if err != nil {
		return pipelines, err
	}

	pipelines = make([]Pipeline, 0)

	for _, namespace := range namespaces {
		namespacePipelines, err := GetPipelinesInNamespace(namespace)
		if err != nil {
			return pipelines, err
		}
		pipelines = append(pipelines, namespacePipelines...)
	}

	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].Identifier < pipelines[j].Identifier
	})

	return pipelines, err
}

func GetPipelinesInNamespace(namespace string) (pipelines []Pipeline, err error) {

	pipelines = nil
	err = nil

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/pipeline?namespace=%s", coreHost, namespace), nil)
	if err != nil {
		return pipelines, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return pipelines, err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode == http.StatusNotFound {
		return pipelines, err
	}

	pips := make([]ScalingFunctions.PipelineIR, 0)
	if err = json.NewDecoder(rsp.Body).Decode(&pips); err != nil {
		return pipelines, err
	}

	pipelines = make([]Pipeline, 0, len(pips))
	for _, p := range pips {
		pipelines = append(pipelines, Pipeline{
			Namespace:  namespace,
			Identifier: p.Identifier,
		})
	}

	return pipelines, err
}

func GetPipeline(namespace, identifier string) (pipeline *ScalingFunctions.PipelineIR, err error) {

	pipeline = nil
	err = nil

	url := fmt.Sprintf("%s/pipeline?namespace=%s&pipeline=%s", coreHost, namespace, identifier)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return pipeline, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return pipeline, err
	}

	pips := make([]ScalingFunctions.PipelineIR, 0)
	decodeErr := json.NewDecoder(rsp.Body).Decode(&pips)
	closeErr := rsp.Body.Close()
	if decodeErr != nil {
		return pipeline, decodeErr
	}
	if closeErr != nil {
		fmt.Print(closeErr)
	}

	for i := range pips {
		if pips[i].Identifier == identifier {
			pipeline = &pips[i]
			return pipeline, err
		}
	}

	err = fmt.Errorf("the pipeline '%s' does not exist in the namespace '%s'", identifier, namespace)
	return pipeline, err
}

func GetRuns() (runs []*run.Run, err error) {

	runs = nil
	err = nil

	namespaces, err := GetNamespaces()
	if err != nil {
		return runs, err
	}

	runs = make([]*run.Run, 0)

	for _, namespace := range namespaces {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/run?namespace=%s", coreHost, namespace), nil)
		if err != nil {
			return runs, err
		}

		rsp, err := client.Do(req)
		if err != nil {
			return runs, err
		}

		r := dataResponse[[]*run.Run]{}
		decodeErr := json.NewDecoder(rsp.Body).Decode(&r)
		closeErr := rsp.Body.Close()

		if rsp.StatusCode == http.StatusNotFound {
			continue
		}
		if decodeErr != nil {
			return runs, decodeErr
		}
		if closeErr != nil {
			fmt.Print(closeErr)
		}

		for _, instance := range r.Data {
			if instance != nil {
				runs = append(runs, instance)
			}
		}
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Id > runs[j].Id
	})

	return runs, err
}

func GetJobs() (jobs []job.Job, err error) {

	jobs = nil
	err = nil

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/job", coreHost), nil)
	if err != nil {
		return jobs, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return jobs, err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return jobs, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := dataResponse[[]job.Job]{}
	err = json.NewDecoder(rsp.Body).Decode(&response)
	if err != nil {
		return jobs, err
	}

	jobs = response.Data
	return jobs, err
}

func CreateJob(job job.Job) (err error) {

	err = nil

	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/job", coreHost), bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", rsp.Status)
	}

	return err
}

func DeleteJob(identifier string) (err error) {

	err = nil

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/job?id=%s", coreHost, identifier), nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", rsp.Status)
	}

	return err
}
