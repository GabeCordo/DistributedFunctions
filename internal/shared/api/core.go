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
	"io"
	"net/http"
	"sort"

	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/component/processor"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/job"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core/database/run"
	"github.com/GabeCordo/ScalingFunctions"
)

func Connect() error {

	url := fmt.Sprintf("http://localhost:5023/gate?=%saction&=connect", "default")

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return nil
}

func Disconnect() error {

	url := fmt.Sprintf("http://localhost:5023/gate?=%saction&=disconnect", "default")
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return nil
}

func Ping() (int, error) {

	rsp, err := http.Get("http://localhost:8136/debug")
	if err != nil {
		return -1, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("bad status: %s", rsp.Status)
	}

	r := Response{}
	err = json.NewDecoder(rsp.Body).Decode(&r)
	if err != nil {
		return -1, err
	}

	if !r.Success {
		return -1, fmt.Errorf("the core did not respond to the ping")
	}

	data, ok := r.Data.(map[string]any)
	if !ok {
		return -1, fmt.Errorf("the ping response is missing data")
	}

	pid, ok := data["pid"].(float64)
	if !ok {
		return -1, fmt.Errorf("the ping response does not contain a pid")
	}

	return int(pid), nil
}

func GetProcessors() ([]processor.Processor, error) {

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8136/processor", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := struct{ Data []processor.Processor }{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func GetModules() ([]processor.ModuleData, error) {

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8136/module", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := struct{ Data []processor.ModuleData }{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

type Function struct {
	Module     string
	Name       string
	Mounted    bool
	Parameters []string
	Returns    []string
}

func GetFunctions() ([]Function, error) {

	modules, err := GetModules()
	if err != nil {
		return nil, err
	}

	functions := make([]Function, 0)

	for _, module := range modules {
		url := fmt.Sprintf("http://localhost:8136/function?module=%s", module.Name)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")

		rsp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		response := struct{ Data []processor.FunctionData }{}
		err = json.NewDecoder(rsp.Body).Decode(&response)
		closeErr := rsp.Body.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			fmt.Print(closeErr)
		}

		for _, f := range response.Data {
			functions = append(functions, Function{
				Module:     module.Name,
				Name:       f.Name,
				Mounted:    f.Mounted,
				Parameters: f.Parameters,
				Returns:    f.Returns,
			})
		}
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	return functions, nil
}

func MountModule(module string) error {
	return setModuleMounted(module, true)
}

func UnmountModule(module string) error {
	return setModuleMounted(module, false)
}

func setModuleMounted(module string, mounted bool) error {

	b, err := json.Marshal(struct {
		Module  string `json:"module"`
		Mounted bool   `json:"mounted"`
	}{module, mounted})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, "http://localhost:8136/module", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return nil
}

func MountFunction(module, function string) error {
	return setFunctionMounted(module, function, true)
}

func UnmountFunction(module, function string) error {
	return setFunctionMounted(module, function, false)
}

func setFunctionMounted(module, function string, mounted bool) error {

	b, err := json.Marshal(struct {
		Module   string `json:"module"`
		Function string `json:"function"`
		Mounted  bool   `json:"mounted"`
	}{module, function, mounted})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, "http://localhost:8136/function", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	}

	return nil
}

func IsPipelineOnCore(host, namespace, pl string) (bool, error) {

	url := fmt.Sprintf("%s/pipeline?namespace=%s", host, namespace)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	pipelines := make([]ScalingFunctions.PipelineIR, 0)
	err = json.NewDecoder(rsp.Body).Decode(&pipelines)
	if err != nil {
		return false, err
	}

	for _, p := range pipelines {
		if p.Identifier == pl {
			return true, nil
		}
	}

	return false, nil
}

func CreatePipelineOnCore(host, namespace string, pl *ScalingFunctions.PipelineIR) error {

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	} else {
		return nil
	}
}

func RemovePipelineFromCore(host, namespace, pipeline string) error {

	url := fmt.Sprintf("%s/pipeline?namespace=%s&pipeline=%s", host, namespace, pipeline)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return errors.New("http status code " + rsp.Status)
	}
	return nil
}

func RunPipelineOnCore(host, namespace, pipeline string) (uint64, error) {

	url := fmt.Sprintf("%s/run", host)

	body := &struct {
		Namespace string            `json:"namespace"`
		Pipeline  string            `json:"pipeline"`
		Metadata  map[string]string `json:"metadata"`
	}{namespace, pipeline, make(map[string]string)}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	response := &struct{ Id uint64 }{}

	if rsp.StatusCode != http.StatusOK {
		return 0, errors.New("http status code " + rsp.Status)
	}

	err = json.NewDecoder(rsp.Body).Decode(response)
	if err != nil {
		return 0, err
	}

	return response.Id, nil
}

func StopRun(host string, id uint64) error {

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
	} else {
		return nil
	}
}

func GetNamespaces() ([]string, error) {

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8136/namespaces", nil)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	response := struct{ Data []string }{}
	decodeErr := json.NewDecoder(rsp.Body).Decode(&response)
	closeErr := rsp.Body.Close()
	if decodeErr != nil {
		return nil, decodeErr
	}
	if closeErr != nil {
		fmt.Print(closeErr)
	}

	return response.Data, nil
}

type Pipeline struct {
	Namespace  string
	Identifier string
}

func GetPipelines() ([]Pipeline, error) {

	namespaces, err := GetNamespaces()
	if err != nil {
		return nil, err
	}

	pipelines := make([]Pipeline, 0)

	for _, namespace := range namespaces {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8136/pipeline?namespace=%s", namespace), nil)
		if err != nil {
			return nil, err
		}

		rsp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		pips := make([]ScalingFunctions.PipelineIR, 0)
		decodeErr := json.NewDecoder(rsp.Body).Decode(&pips)
		closeErr := rsp.Body.Close()

		// the namespace does not have any pipelines registered on the core
		if rsp.StatusCode == http.StatusNotFound {
			continue
		}
		if decodeErr != nil {
			return nil, decodeErr
		}
		if closeErr != nil {
			fmt.Print(closeErr)
		}

		for _, p := range pips {
			pipelines = append(pipelines, Pipeline{
				Namespace:  namespace,
				Identifier: p.Identifier,
			})
		}
	}

	sort.Slice(pipelines, func(i, j int) bool {
		return pipelines[i].Identifier < pipelines[j].Identifier
	})

	return pipelines, nil
}

func GetPipeline(namespace, identifier string) (*ScalingFunctions.PipelineIR, error) {

	url := fmt.Sprintf("http://localhost:8136/pipeline?namespace=%s&pipeline=%s", namespace, identifier)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	pips := make([]ScalingFunctions.PipelineIR, 0)
	decodeErr := json.NewDecoder(rsp.Body).Decode(&pips)
	closeErr := rsp.Body.Close()
	if decodeErr != nil {
		return nil, decodeErr
	}
	if closeErr != nil {
		fmt.Print(closeErr)
	}

	// the endpoint does not filter on the pipeline identifier, match it locally
	for i := range pips {
		if pips[i].Identifier == identifier {
			return &pips[i], nil
		}
	}

	return nil, fmt.Errorf("the pipeline '%s' does not exist in the namespace '%s'", identifier, namespace)
}

func GetRuns() ([]*run.Run, error) {

	namespaces, err := GetNamespaces()
	if err != nil {
		return nil, err
	}

	runs := make([]*run.Run, 0)

	for _, namespace := range namespaces {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8136/run?namespace=%s", namespace), nil)
		if err != nil {
			return nil, err
		}

		rsp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		r := struct{ Data []*run.Run }{}
		decodeErr := json.NewDecoder(rsp.Body).Decode(&r)
		closeErr := rsp.Body.Close()

		// the namespace does not have any runs tracked by the core
		if rsp.StatusCode == http.StatusNotFound {
			continue
		}
		if decodeErr != nil {
			return nil, decodeErr
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

	return runs, nil
}

func GetJobs() ([]job.Job, error) {

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8136/job", nil)
	if err != nil {
		return nil, err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", rsp.Status)
	}

	response := struct{ Data []job.Job }{}
	if err = json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func CreateJob(job job.Job) error {

	b, err := json.Marshal(job)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8136/job", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	} else {
		return nil
	}
}

func DeleteJob(identifier string) error {

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:8136/job?id=%s", identifier), nil)
	if err != nil {
		return err
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Print(err)
		}
	}(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", rsp.Status)
	} else {
		return nil
	}
}
