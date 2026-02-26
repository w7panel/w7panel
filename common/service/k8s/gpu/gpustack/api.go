package gpustack

import (
	"encoding/json"
	"errors"
	"strconv"

	"gitee.com/we7coreteam/k8s-offline/common/helper"
)

type gpustackApi struct {
	gpuStackApiUrl string
	userName       string
	password       string
}

func NewGpuStackApi(gpuStackApiUrl, userName, password string) *gpustackApi {
	return &gpustackApi{
		gpuStackApiUrl: gpuStackApiUrl,
		userName:       userName,
		password:       password,
	}
}

type WorkerResponse struct {
	Items []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		// Other fields from the response can be added here
	} `json:"items"`
}

func (g *gpustackApi) GetGpuStackWorkers() (map[string]int, error) {
	url := g.gpuStackApiUrl + "/v1/workers"
	response, err := helper.RetryHttpClient().R().SetBasicAuth(g.userName, g.password).Get(url)
	if err != nil {
		return nil, err
	}
	if !response.IsSuccess() {
		return nil, errors.New("GetGpuStackWorkers error: " + response.String())
	}

	var workers WorkerResponse
	if err := json.Unmarshal(response.Body(), &workers); err != nil {
		return nil, err
	}

	// Create name:id map
	nameIDMap := make(map[string]int)
	for _, worker := range workers.Items {
		nameIDMap[worker.Name] = worker.ID
	}
	return nameIDMap, nil
}

func (g *gpustackApi) DeleteGpuStackWorkerByName(workerName string) error {

	workerIds, err := g.GetGpuStackWorkers()
	if err != nil {
		return err
	}
	if _, ok := workerIds[workerName]; !ok {
		return errors.New("worker not found")
	}
	url := g.gpuStackApiUrl + "/v1/workers/" + strconv.Itoa(workerIds[workerName])
	response, err := helper.RetryHttpClient().R().SetBasicAuth(g.userName, g.password).Delete(url)
	if err != nil {
		return err
	}
	if !response.IsSuccess() {
		return errors.New("DeleteGpuStackWorkerByName error: " + response.String())
	}
	return nil

}
