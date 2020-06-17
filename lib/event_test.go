package lib

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"strings"
	"sync"
	"testing"
)

func TestEvents(t *testing.T) {
	err := LoadConfig("../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	Config.PgConn, err = docker.Postgres(ctx, &wg)
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, requests := mocks.CamundaServer(ctx, &wg)
	Config.ProcessEngineUrl = camundaUrl

	err = InitEventSourcing(mocks.Kafka())
	if err != nil {
		t.Error(err)
		return
	}

	defer CloseEventSourcing()

	t.Run("publish version 1 deployment", publishVersion1Deployment("1", "testname", bpmnExample, svgExample))

	t.Run("check version 1 camunda request", checkCamundaRequest(requests, "testname", bpmnExample, svgExample))

	t.Run("publish version 1 invalid deployment", publishVersion1Deployment("1", "testname", "invalid", svgExample))

	t.Run("check version 1 invalid camunda request", checkCamundaRequest(requests, "testname", createBlankProcess(), createBlankSvg()))

	t.Run("publish version 2 deployment", publishVersion2Deployment("1", "testname", bpmnExample, svgExample))

	t.Run("check version 2 camunda request", checkCamundaRequest(requests, "testname", bpmnExample, svgExample))

	t.Run("publish version 2 invalid deployment", publishVersion2Deployment("1", "testname", "invalid", svgExample))

	t.Run("check version 2 invalid camunda request", checkCamundaRequest(requests, "testname", createBlankProcess(), createBlankSvg()))

	t.Run("publish explicit version 1 deployment", publishExplVersion1Deployment("1", "3", "testname", bpmnExample, svgExample))

	t.Run("check explicit version 1 camunda request", checkCamundaRequest(requests, "testname", bpmnExample, svgExample))

	t.Run("publish explicit version 1 invalid deployment", publishExplVersion1Deployment("1", "3", "testname", "invalid", svgExample))

	t.Run("check explicit version 1 invalid camunda request", checkCamundaRequest(requests, "testname", createBlankProcess(), createBlankSvg()))

	t.Run("publish explicit version '' deployment", publishExplVersion1Deployment("", "3", "testname", bpmnExample, svgExample))

	t.Run("check explicit version '' camunda request", checkCamundaRequest(requests, "testname", bpmnExample, svgExample))

	t.Run("publish explicit version '' invalid deployment", publishExplVersion1Deployment("", "3", "testname", "invalid", svgExample))

	t.Run("check explicit version '' invalid camunda request", checkCamundaRequest(requests, "testname", createBlankProcess(), createBlankSvg()))
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

func publishVersion1Deployment(id string, name string, xml string, svg string) func(t *testing.T) {
	return func(t *testing.T) {
		msg, err := json.Marshal(DeploymentCommand{
			Command: "PUT",
			Id:      id,
			Owner:   "test",
			Deployment: map[string]interface{}{
				"id":   id,
				"xml":  xml,
				"svg":  svg,
				"name": name,
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = cqrs.Publish(Config.DeploymentTopic, id, msg)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func publishExplVersion1Deployment(version string, id string, name string, xml string, svg string) func(t *testing.T) {
	return func(t *testing.T) {
		msg, err := json.Marshal(DeploymentCommand{
			Command: "PUT",
			Id:      id,
			Owner:   "test",
			Deployment: map[string]interface{}{
				"id":      id,
				"version": version,
				"xml":     xml,
				"svg":     svg,
				"name":    name,
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = cqrs.Publish(Config.DeploymentTopic, id, msg)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func publishVersion2Deployment(id string, name string, xml string, svg string) func(t *testing.T) {
	return func(t *testing.T) {
		msg, err := json.Marshal(DeploymentCommand{
			Command: "PUT",
			Id:      id,
			Owner:   "test",
			Deployment: map[string]interface{}{
				"version": "2",
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
		err = cqrs.Publish(Config.DeploymentTopic, id, msg)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
