package tests

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/messages"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/helper"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"strings"
	"sync"
	"testing"
)

func TestEvents(t *testing.T) {
	config, err := configuration.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	config.ShardingDb, err = docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}
	config.WrapperDb = config.ShardingDb

	camundaUrl, requests := mocks.CamundaServer(ctx, &wg)
	s, err := shards.New(config.ShardingDb, cache.None)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		t.Error(err)
		return
	}

	v, err := vid.New(config.WrapperDb)
	if err != nil {
		t.Error(err)
		return
	}

	c := camunda.New(config, v, s, nil)

	e, err := events.New(config, mocks.Kafka(), v, c, nil)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("sub test naming schema is deprecated. only one deployment version is supported for this service")

	t.Run("publish version 1 deployment", publishDeployment(config, e, "1", "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("check version 1 camunda request", checkCamundaRequest(requests, "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("publish version 1 invalid deployment", publishDeployment(config, e, "1", "testname", "invalid", helper.SvgExample))

	t.Run("check version 1 invalid camunda request", checkCamundaRequest(requests, "testname", camunda.CreateBlankProcess(), camunda.CreateBlankSvg()))

	t.Run("publish version 2 deployment", publishDeployment(config, e, "1", "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("check version 2 camunda request", checkCamundaRequest(requests, "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("publish version 2 invalid deployment", publishDeployment(config, e, "1", "testname", "invalid", helper.SvgExample))

	t.Run("check version 2 invalid camunda request", checkCamundaRequest(requests, "testname", camunda.CreateBlankProcess(), camunda.CreateBlankSvg()))

	t.Run("publish explicit version 1 deployment", publishDeployment(config, e, "3", "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("check explicit version 1 camunda request", checkCamundaRequest(requests, "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("publish explicit version 1 invalid deployment", publishDeployment(config, e, "3", "testname", "invalid", helper.SvgExample))

	t.Run("check explicit version 1 invalid camunda request", checkCamundaRequest(requests, "testname", camunda.CreateBlankProcess(), camunda.CreateBlankSvg()))

	t.Run("publish explicit version '' deployment", publishDeployment(config, e, "3", "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("check explicit version '' camunda request", checkCamundaRequest(requests, "testname", helper.BpmnExample, helper.SvgExample))

	t.Run("publish explicit version '' invalid deployment", publishDeployment(config, e, "3", "testname", "invalid", helper.SvgExample))

	t.Run("check explicit version '' invalid camunda request", checkCamundaRequest(requests, "testname", camunda.CreateBlankProcess(), camunda.CreateBlankSvg()))
}

func checkCamundaRequest(requests chan mocks.Request, expectedName string, expectedXml string, expectedSvg string) func(t *testing.T) {
	return func(t *testing.T) {
		if len(requests) != 1 {
			t.Error(len(requests))
			return
		}
		req := <-requests
		if req.Url != "/engine-rest/deployment/create" {
			t.Error(req.Url)
		}

		xml, svg, name, user, err := parseCamundaMessage(req.Payload)
		if err != nil {
			t.Error(err)
			return
		}

		if removeProcessId(xml) != removeProcessId(expectedXml) {
			t.Error("got: ", xml, "\n", "expected:", expectedXml)
		}
		if svg != expectedSvg {
			t.Error("got: ", svg, "\n", "expected:", expectedSvg)
		}
		if name != expectedName {
			t.Error("got: ", name, "\n", "expected:", expectedName)
		}
		if user != "test" {
			t.Error("got: ", user, "\n", "expected:", "test")
		}
	}
}

func removeProcessId(xml string) string {
	pattern := "<bpmn:process id='"
	i := strings.Index(xml, pattern)
	if i == -1 {
		return xml
	} else {
		i = i + len(pattern)
	}
	j := strings.Index(xml[i:], "'")
	result := xml[:i] + xml[i+j:]
	return result
}

func parseCamundaMessage(payload []byte) (xml string, svg string, name string, user string, err error) {
	cleanPl := strings.ReplaceAll(string(payload), "\r", "")
	form := map[string]string{}
	for _, segment := range strings.Split(cleanPl, "-----------------------------") {
		if segment != "" {
			segmentParts := strings.Split(segment, "\n\n")
			headers := segmentParts[0]
			headersList := strings.Split(headers, "\n")
			for _, header := range headersList {
				headerParts := strings.Split(header, ":")
				headerName := headerParts[0]
				if headerName == "Content-Disposition" {
					headerBodySegments := strings.Split(headerParts[1], ";")
					for _, headerBodySegment := range headerBodySegments {
						headerBodySegmentParts := strings.Split(headerBodySegment, "=")
						if strings.TrimSpace(headerBodySegmentParts[0]) == "name" {
							body := segmentParts[1]
							form[strings.TrimSpace(headerBodySegmentParts[1])] = strings.TrimSpace(body)
						}
					}
				}
			}
		}
	}

	xml = form[`"data"`]
	svg = form[`"diagram"`]
	name = form[`"deployment-name"`]
	user = form[`"tenant-id"`]
	return
}

type TestDeploymentCommand struct {
	Version    int64       `json:"version"`
	Command    string      `json:"command"`
	Id         string      `json:"id"`
	Owner      string      `json:"owner"`
	Deployment interface{} `json:"deployment"`
}

func publishDeployment(config configuration.Config, events *events.Events, id string, name string, xml string, svg string) func(t *testing.T) {
	return func(t *testing.T) {
		msg, err := json.Marshal(TestDeploymentCommand{
			Version: messages.CurrentVersion,
			Command: "PUT",
			Id:      id,
			Owner:   "test",
			Deployment: map[string]interface{}{
				"version": messages.CurrentVersion,
				"id":      id,
				"name":    name,
				"diagram": map[string]interface{}{
					"xml_deployed": xml,
					"svg":          svg,
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}

		err = events.GetPublisher().Publish(config.DeploymentTopic, id, msg)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
