package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda/model"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestGetParameter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config, wrapperUrl, _, e, err := helper.CreateTestEnv(ctx, &wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	deploymentId := "withInput"
	t.Run("deploy process with input", testDeployProcessWithInput(e, deploymentId, processWithInput))

	idWithForm := "withForm"
	t.Run("deploy process with form", testDeployProcessWithInput(e, idWithForm, processWithForm))

	t.Run("check parameter of normal process", checkProcessParameterDeclaration(wrapperUrl, deploymentId, map[string]model.Variable{}))
	t.Run("check parameter of form process", checkProcessParameterDeclaration(wrapperUrl, idWithForm, map[string]model.Variable{"inputTemperature": {
		Value:     nil,
		Type:      "Long",
		ValueInfo: map[string]interface{}{},
	}}))
}

func TestStartWithInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config, wrapperUrl, shard, e, err := helper.CreateTestEnv(ctx, &wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	deploymentId := "withInput"
	t.Run("deploy process with input", testDeployProcessWithInput(e, deploymentId, processWithInput))

	t.Run("start process with input inputTemperature 30", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 30}))
	t.Run("check and finish task with temp 30", testCheckProcessWithInputTask(shard, float64(30)))

	t.Run("start process with input inputTemperature 21", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 21}))
	t.Run("check and finish task with temp 21", testCheckProcessWithInputTask(shard, float64(21)))

	t.Run("start process with partially unused inputs", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": 10, "unused": "foo"}))
	t.Run("check and finish task with partially unused inputs", testCheckProcessWithInputTask(shard, float64(10)))

	t.Run("start process with input inputTemperature 30 as string", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": "30"}))
	t.Run("check and finish task with temp 30 as string", testCheckProcessWithInputTask(shard, "30"))

	t.Run("start process with input inputTemperature 21 as string", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": "21"}))
	t.Run("check and finish task with temp 21 as string", testCheckProcessWithInputTask(shard, "21"))

	t.Run("start process with partially unused inputs as string", testStartProcessWithInput(wrapperUrl, deploymentId, map[string]interface{}{"inputTemperature": "10", "unused": 13}))
	t.Run("check and finish task with partially unused inputs as string", testCheckProcessWithInputTask(shard, "10"))

	t.Run("start process with input inputTemperature 30 as raw string", testStartProcessWithRawStringInput(wrapperUrl, deploymentId, map[string]string{"inputTemperature": "30"}))
	t.Run("check and finish task with temp 30 as raw string", testCheckProcessWithInputTask(shard, float64(30)))

	t.Run("start process with input inputTemperature 21 as raw string", testStartProcessWithRawStringInput(wrapperUrl, deploymentId, map[string]string{"inputTemperature": "21"}))
	t.Run("check and finish task with temp 21 as raw string", testCheckProcessWithInputTask(shard, float64(21)))

	t.Run("start process with partially unused inputs as raw string", testStartProcessWithRawStringInput(wrapperUrl, deploymentId, map[string]string{"inputTemperature": "10", "unused": "batz"}))
	t.Run("check and finish task with partially unused inputs as raw string", testCheckProcessWithInputTask(shard, float64(10)))

	t.Run("start process with str inputs as raw string", testStartProcessWithRawStringInput(wrapperUrl, deploymentId, map[string]string{"inputTemperature": "foo bar batz"}))
	t.Run("check and finish task with str inputs as raw string", testCheckProcessWithInputTask(shard, "foo bar batz"))

	t.Run("deploy process without input", testDeployProcessWithInput(e, "withoutInput", processWithoutInput))
	t.Run("start process without inputs", testStartProcessWithInput(wrapperUrl, "withoutInput", nil))
	t.Run("check and finish task with temp 42", testCheckProcessWithInputTask(shard, "42"))
}

func TestStartWithInputForm(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()

	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config, wrapperUrl, shard, e, err := helper.CreateTestEnv(ctx, &wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	idWithForm := "withForm"

	t.Run("deploy process with form", testDeployProcessWithInput(e, idWithForm, processWithForm))

	t.Run("start process with input inputTemperature 30", testStartProcessWithInput(wrapperUrl, idWithForm, map[string]interface{}{"inputTemperature": 30}))
	t.Run("check and finish task with temp 30", testCheckProcessWithInputTask(shard, float64(30)))

	t.Run("start process with input inputTemperature 21", testStartProcessWithInput(wrapperUrl, idWithForm, map[string]interface{}{"inputTemperature": 21}))
	t.Run("check and finish task with temp 21", testCheckProcessWithInputTask(shard, float64(21)))

	t.Run("start process with partially unused inputs", testStartProcessWithInput(wrapperUrl, idWithForm, map[string]interface{}{"inputTemperature": 42, "unused": "foo"}))
	t.Run("check and finish task with partially unused inputs", testCheckProcessWithInputTask(shard, float64(42)))

	t.Run("start process with input inputTemperature 30 as raw string", testStartProcessWithRawStringInput(wrapperUrl, idWithForm, map[string]string{"inputTemperature": "30"}))
	t.Run("check and finish task with temp 30 as raw string", testCheckProcessWithInputTask(shard, float64(30)))

	t.Run("start process with input inputTemperature 21 as raw string", testStartProcessWithRawStringInput(wrapperUrl, idWithForm, map[string]string{"inputTemperature": "21"}))
	t.Run("check and finish task with temp 21 as raw string", testCheckProcessWithInputTask(shard, float64(21)))

	t.Run("start process with partially unused inputs as raw string", testStartProcessWithRawStringInput(wrapperUrl, idWithForm, map[string]string{"inputTemperature": "42", "unused": "batz"}))
	t.Run("check and finish task with partially unused inputs as raw string", testCheckProcessWithInputTask(shard, float64(42)))
}

func testCheckProcessWithInputTask(url string, expected interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		tasks, err := fetchTestTask(url)
		if err != nil {
			t.Error(err)
			return
		}
		if len(tasks) != 1 {
			t.Error(len(tasks), tasks)
			return
		}
		inputs := tasks[0].Variables["inputs"].Value
		if !reflect.DeepEqual(inputs, expected) {
			t.Error(reflect.TypeOf(inputs), inputs, reflect.TypeOf(expected), expected)
			return
		}
		err = completeTask(url, tasks[0].Id)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testStartProcessWithInput(wrapper string, id string, inputs map[string]interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		values := url.Values{}
		for key, value := range inputs {
			val, err := json.Marshal(value)
			if err != nil {
				t.Error(err)
				return
			}
			values.Add(key, string(val))
		}

		req, err := http.NewRequest("GET", wrapper+"/deployment/"+id+"/start", nil)
		if inputs != nil {
			req, err = http.NewRequest("GET", wrapper+"/deployment/"+id+"/start?"+values.Encode(), nil)
		}
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", string(helper.Jwt))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			temp, _ := ioutil.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}
	}
}

func testStartProcessWithRawStringInput(wrapper string, id string, inputs map[string]string) func(t *testing.T) {
	return func(t *testing.T) {
		values := url.Values{}
		for key, value := range inputs {
			values.Add(key, value)
		}

		req, err := http.NewRequest("GET", wrapper+"/deployment/"+id+"/start", nil)
		if inputs != nil {
			req, err = http.NewRequest("GET", wrapper+"/deployment/"+id+"/start?"+values.Encode(), nil)
		}
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", string(helper.Jwt))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			temp, _ := ioutil.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}
	}
}

func checkProcessParameterDeclaration(wrapper string, id string, expected map[string]model.Variable) func(t *testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest("GET", wrapper+"/deployment/"+id+"/parameter", nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", string(helper.Jwt))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			temp, _ := ioutil.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}
		result := map[string]model.Variable{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, expected) {
			t.Error(result, expected)
			return
		}
	}
}

func testDeployProcessWithInput(e *events.Events, id string, bpmn string) func(t *testing.T) {
	return func(t *testing.T) {
		err := e.HandleDeploymentCreate(helper.JwtPayload.UserId, id, "processWithInput", bpmn, helper.SvgExample, "")
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func fetchTestTask(shard string) (tasks []CamundaExternalTask, err error) {
	fetchRequest := CamundaFetchRequest{
		WorkerId: "test",
		MaxTasks: 10,
		Topics:   []CamundaTopic{{LockDuration: 1, Name: "optimistic"}},
	}
	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(fetchRequest)
	if err != nil {
		return
	}
	endpoint := shard + "/engine-rest/external-task/fetchAndLock"
	resp, err := client.Post(endpoint, "application/json", b)
	if err != nil {
		return tasks, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		temp, err := ioutil.ReadAll(resp.Body)
		err = errors.New(fmt.Sprintln(endpoint, resp.Status, resp.StatusCode, string(temp), err))
		return tasks, err
	}
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	return
}

func completeTask(shard string, taskId string) (err error) {
	var completeRequest CamundaCompleteRequest

	completeRequest = CamundaCompleteRequest{WorkerId: "test"}

	log.Println("Start complete Request")
	client := http.Client{Timeout: 5 * time.Second}
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(completeRequest)
	if err != nil {
		return
	}
	resp, err := client.Post(shard+"/engine-rest/external-task/"+taskId+"/complete", "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	return nil
}

type CamundaCompleteRequest struct {
	WorkerId string `json:"workerId,omitempty"`
}

type CamundaFetchRequest struct {
	WorkerId string         `json:"workerId,omitempty"`
	MaxTasks int64          `json:"maxTasks,omitempty"`
	Topics   []CamundaTopic `json:"topics,omitempty"`
}

type CamundaTopic struct {
	Name         string `json:"topicName,omitempty"`
	LockDuration int64  `json:"lockDuration,omitempty"`
}

type CamundaExternalTask struct {
	Id                  string                     `json:"id,omitempty"`
	Variables           map[string]CamundaVariable `json:"variables,omitempty"`
	ActivityId          string                     `json:"activityId,omitempty"`
	Retries             int64                      `json:"retries"`
	ExecutionId         string                     `json:"executionId"`
	ProcessInstanceId   string                     `json:"processInstanceId"`
	ProcessDefinitionId string                     `json:"processDefinitionId"`
	TenantId            string                     `json:"tenantId"`
	Error               string                     `json:"errorMessage"`
}

type CamundaVariable struct {
	Type  string      `json:"type,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

const processWithInput = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:camunda="http://camunda.org/schema/1.0/bpmn" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="moses_termostat_test" isExecutable="true" name="moses_termostat_test">
    <bpmn:startEvent id="StartEvent_1">
      <bpmn:outgoing>SequenceFlow_1b16ztv</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:sequenceFlow id="SequenceFlow_1b16ztv" sourceRef="StartEvent_1" targetRef="Task_1ol5jfc"/>
    <bpmn:endEvent id="EndEvent_1oihco2">
      <bpmn:incoming>SequenceFlow_0qjgxu7</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="SequenceFlow_0qjgxu7" sourceRef="Task_1ol5jfc" targetRef="EndEvent_1oihco2"/>
    <bpmn:serviceTask id="Task_1ol5jfc" name="Thermostat setTemperatureFunction" camunda:type="external" camunda:topic="optimistic">
      <bpmn:extensionElements>
        <camunda:inputOutput>
          <camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
	"device_class": {
		"id": "urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393",
		"name": "",
		"image": "",
		"rdf_type": ""
	},
	"device_id": "urn:infai:ses:device:5be329b2-7675-4b4e-b6b0-cd954e2eeeed",
	"service_id": "urn:infai:ses:service:e7e7da30-d1da-44e6-9af1-2b450fa5a333",
	"input": 0
}]]></camunda:inputParameter>
          <camunda:inputParameter name="inputs">${inputTemperature}</camunda:inputParameter>
        </camunda:inputOutput>
      </bpmn:extensionElements>
      <bpmn:incoming>SequenceFlow_1b16ztv</bpmn:incoming>
      <bpmn:outgoing>SequenceFlow_0qjgxu7</bpmn:outgoing>
    </bpmn:serviceTask>
  </bpmn:process>
</bpmn:definitions>`

const processWithoutInput = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:camunda="http://camunda.org/schema/1.0/bpmn" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn"><bpmn:process id="moses_termostat_test" isExecutable="true" name="moses_termostat_test"><bpmn:startEvent id="StartEvent_1"><bpmn:outgoing>SequenceFlow_1b16ztv</bpmn:outgoing></bpmn:startEvent><bpmn:sequenceFlow id="SequenceFlow_1b16ztv" sourceRef="StartEvent_1" targetRef="Task_1ol5jfc"/><bpmn:endEvent id="EndEvent_1oihco2"><bpmn:incoming>SequenceFlow_0qjgxu7</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id="SequenceFlow_0qjgxu7" sourceRef="Task_1ol5jfc" targetRef="EndEvent_1oihco2"/><bpmn:serviceTask id="Task_1ol5jfc" name="Thermostat setTemperatureFunction" camunda:type="external" camunda:topic="optimistic"><bpmn:extensionElements><camunda:inputOutput><camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
	"device_class": {
		"id": "urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393",
		"name": "",
		"image": "",
		"rdf_type": ""
	},
	"device_id": "urn:infai:ses:device:5be329b2-7675-4b4e-b6b0-cd954e2eeeed",
	"service_id": "urn:infai:ses:service:e7e7da30-d1da-44e6-9af1-2b450fa5a333",
	"input": 0
}]]></camunda:inputParameter><camunda:inputParameter name="inputs">42</camunda:inputParameter></camunda:inputOutput></bpmn:extensionElements><bpmn:incoming>SequenceFlow_1b16ztv</bpmn:incoming><bpmn:outgoing>SequenceFlow_0qjgxu7</bpmn:outgoing></bpmn:serviceTask></bpmn:process><bpmndi:BPMNDiagram id="BPMNDiagram_1"><bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="moses_termostat_test"><bpmndi:BPMNShape id="_BPMNShape_StartEvent_2" bpmnElement="StartEvent_1"><dc:Bounds x="173" y="102" width="36" height="36"/></bpmndi:BPMNShape><bpmndi:BPMNEdge id="SequenceFlow_1b16ztv_di" bpmnElement="SequenceFlow_1b16ztv"><di:waypoint x="209" y="120"/><di:waypoint x="260" y="120"/></bpmndi:BPMNEdge><bpmndi:BPMNShape id="EndEvent_1oihco2_di" bpmnElement="EndEvent_1oihco2"><dc:Bounds x="412" y="102" width="36" height="36"/></bpmndi:BPMNShape><bpmndi:BPMNEdge id="SequenceFlow_0qjgxu7_di" bpmnElement="SequenceFlow_0qjgxu7"><di:waypoint x="360" y="120"/><di:waypoint x="412" y="120"/></bpmndi:BPMNEdge><bpmndi:BPMNShape id="ServiceTask_1is4vpl_di" bpmnElement="Task_1ol5jfc"><dc:Bounds x="260" y="80" width="100" height="80"/></bpmndi:BPMNShape></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`

const processWithForm = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" xmlns:camunda="http://camunda.org/schema/1.0/bpmn" xmlns:di="http://www.omg.org/spec/DD/20100524/DI" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="moses_termostat_test" isExecutable="true" name="moses_termostat_test">
    <bpmn:startEvent id="StartEvent_1">
      <bpmn:extensionElements>
        <camunda:formData>
          <camunda:formField id="inputTemperature" label="inputTemperature" type="long">
            <camunda:validation>
              <camunda:constraint name="min" config="20"/>
            </camunda:validation>
          </camunda:formField>
        </camunda:formData>
      </bpmn:extensionElements>
      <bpmn:outgoing>SequenceFlow_1b16ztv</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:sequenceFlow id="SequenceFlow_1b16ztv" sourceRef="StartEvent_1" targetRef="Task_1ol5jfc"/>
    <bpmn:endEvent id="EndEvent_1oihco2">
      <bpmn:incoming>SequenceFlow_0qjgxu7</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="SequenceFlow_0qjgxu7" sourceRef="Task_1ol5jfc" targetRef="EndEvent_1oihco2"/>
    <bpmn:serviceTask id="Task_1ol5jfc" name="Thermostat setTemperatureFunction" camunda:type="external" camunda:topic="optimistic">
      <bpmn:extensionElements>
        <camunda:inputOutput>
          <camunda:inputParameter name="payload"><![CDATA[{
	"function": {
		"id": "urn:infai:ses:controlling-function:99240d90-02dd-4d4f-a47c-069cfe77629c",
		"name": "",
		"concept_id": "",
		"rdf_type": ""
	},
	"characteristic_id": "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
	"device_class": {
		"id": "urn:infai:ses:device-class:997937d6-c5f3-4486-b67c-114675038393",
		"name": "",
		"image": "",
		"rdf_type": ""
	},
	"device_id": "urn:infai:ses:device:5be329b2-7675-4b4e-b6b0-cd954e2eeeed",
	"service_id": "urn:infai:ses:service:e7e7da30-d1da-44e6-9af1-2b450fa5a333",
	"input": 0
}]]></camunda:inputParameter>
          <camunda:inputParameter name="inputs">${inputTemperature}</camunda:inputParameter>
        </camunda:inputOutput>
      </bpmn:extensionElements>
      <bpmn:incoming>SequenceFlow_1b16ztv</bpmn:incoming>
      <bpmn:outgoing>SequenceFlow_0qjgxu7</bpmn:outgoing>
    </bpmn:serviceTask>
  </bpmn:process>
</bpmn:definitions>`
