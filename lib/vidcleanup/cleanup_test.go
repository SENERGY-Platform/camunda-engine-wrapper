package vidcleanup

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/cache"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/shards"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/docker"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/tests/mocks"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

const bpmnExample = `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' xmlns:di='http://www.omg.org/spec/DD/20100524/DI' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='Process_1' isExecutable='true'><bpmn:startEvent id='StartEvent_1'><bpmn:outgoing>SequenceFlow_02ibfc0</bpmn:outgoing></bpmn:startEvent><bpmn:endEvent id='EndEvent_1728wjv'><bpmn:incoming>SequenceFlow_02ibfc0</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id='SequenceFlow_02ibfc0' sourceRef='StartEvent_1' targetRef='EndEvent_1728wjv'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='Process_1'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNShape id='EndEvent_1728wjv_di' bpmnElement='EndEvent_1728wjv'><dc:Bounds x='259' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNEdge id='SequenceFlow_02ibfc0_di' bpmnElement='SequenceFlow_02ibfc0'><di:waypoint x='209' y='120'></di:waypoint><di:waypoint x='259' y='120'></di:waypoint></bpmndi:BPMNEdge></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
const svgExample = `<svg height='48' version='1.1' viewBox='167 96 134 48' width='134' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'><defs><marker id='sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p' markerHeight='10' markerWidth='10' orient='auto' refX='11' refY='10' viewBox='0 0 20 20'><path d='M 1 5 L 11 10 L 1 15 Z' style='fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;'/></marker></defs><g class='djs-group'><g class='djs-element djs-connection' data-element-id='SequenceFlow_04zz9eb' style='display: block;'><g class='djs-visual'><path d='m  209,120L259,120 ' style='fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url(&apos;#sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p&apos;);'/></g><polyline class='djs-hit' points='209,120 259,120 ' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;'/><rect class='djs-outline' height='12' style='fill: none;' width='62' x='203' y='114'/></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='StartEvent_1' style='display: block;' transform='translate(173 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;'/><path d='m 8.459999999999999,11.34 l 0,12.6 l 18.900000000000002,0 l 0,-12.6 z l 9.450000000000001,5.4 l 9.450000000000001,-5.4' style='fill: white; stroke-width: 1px; stroke: black;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='EndEvent_056p30q' style='display: block;' transform='translate(259 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g></svg>`

func TestClearUnlinkedDeployments(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	defer wg.Wait()
	defer cancel()

	pgConn, err := docker.Postgres(ctx, &wg, "test")
	if err != nil {
		t.Error(err)
		return
	}

	v, err := vid.New(pgConn)
	if err != nil {
		t.Error(err)
		return
	}

	s, err := shards.New(pgConn, cache.None)
	if err != nil {
		t.Error(err)
		return
	}

	_, camundaPgIp, camundaPgPort, err := docker.PostgresWithNetwork(ctx, &wg, "camunda")
	if err != nil {
		t.Error(err)
		return
	}

	camundaUrl, err := docker.Camunda(ctx, &wg, camundaPgIp, camundaPgPort)
	if err != nil {
		t.Error(err)
		return
	}
	err = s.EnsureShard(camundaUrl)
	if err != nil {
		t.Error(err)
		return
	}

	expectedProcessId := ""
	deletedProcessId := ""

	t.Run("create normal process", testCreateNormalProcess(camundaUrl, v, "expectedVid", &expectedProcessId))
	t.Run("create missing vid", testCreateProcess(camundaUrl, &deletedProcessId))
	t.Run("create missing process", testCreateVid(v, "missingProcessVid", "missingProcess"))

	t.Run("check camunda process count", testCheckCamundaProcessCount(camundaUrl, 2))
	t.Run("check false vid", testCheckVid(v, "missingProcessVid", "missingProcess", true))
	t.Run("check true vid", testCheckVid(v, "expectedVid", expectedProcessId, true))
	t.Run("check camunda process", testCheckCamundaProcess(camundaUrl, expectedProcessId, true))
	t.Run("check camunda deleted process", testCheckCamundaProcess(camundaUrl, deletedProcessId, true))

	cqrs := mocks.Kafka()
	t.Run("cleanup", testCleanup(pgConn, cqrs))

	t.Run("check camunda process count after cleanup", testCheckCamundaProcessCount(camundaUrl, 1))

	t.Run("check true vid after cleanup", testCheckVid(v, "expectedVid", expectedProcessId, true))
	t.Run("check camunda process after cleanup", testCheckCamundaProcess(camundaUrl, expectedProcessId, true))
	t.Run("check camunda deleted process after cleanup", testCheckCamundaProcess(camundaUrl, deletedProcessId, false))

	//vid will be deleted by event so testCheckVid(v, "missingProcessVid", "missingProcess", false)) wil not work here
	t.Run("check kafka delete messages", testCheckKafkaDeletes(cqrs, "missingProcessVid"))
}

func testCheckKafkaDeletes(cqrs *mocks.KafkaMock, expectedDeletes ...string) func(t *testing.T) {
	return func(t *testing.T) {
		if cqrs == nil {
			t.Fatal("missing kafka mock")
		}
		actualDeletes := cqrs.Produced["deployment"]
		if len(expectedDeletes) != len(actualDeletes) {
			t.Error(actualDeletes, expectedDeletes)
			return
		}
		for i, del := range actualDeletes {
			msg := DeploymentDeleteCommand{}
			err := json.Unmarshal([]byte(del), &msg)
			if err != nil {
				t.Error(err)
				return
			}
			if msg.Command != "DELETE" {
				t.Error(del)
				return
			}
			if msg.Id != expectedDeletes[i] {
				t.Error(msg.Id, expectedDeletes[i], del)
				return
			}
		}
	}
}

func testCleanup(conn string, cqrs *mocks.KafkaMock) func(t *testing.T) {
	return func(t *testing.T) {
		err := ClearUnlinkedDeployments(conn, "deployment", cqrs, 0)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testCheckCamundaProcess(shard string, pid string, expectExistence bool) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := http.Get(shard + "/engine-rest/deployment/" + url.PathEscape(pid))
		if err != nil {
			t.Error(err)
			return
		}
		if (resp.StatusCode == 200) != expectExistence {
			t.Error(resp.StatusCode)
			return
		}
	}
}

func testCheckVid(v *vid.Vid, vid string, pid string, expectExistence bool) func(t *testing.T) {
	return func(t *testing.T) {
		vidExists, err := v.VidExists(vid)
		if err != nil {
			t.Error(err)
			return
		}
		if vidExists != expectExistence {
			t.Error(vidExists, expectExistence)
			return
		}
		deplId, exists, err := v.GetDeploymentId(vid)
		if err != nil {
			t.Error(err)
			return
		}
		if exists != expectExistence {
			t.Error(vidExists, expectExistence)
			return
		}
		if exists && deplId != pid {
			t.Error(deplId, pid)
			return
		}
	}
}

type Count struct {
	Count int `json:"count"`
}

func testCheckCamundaProcessCount(url string, expected int) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := http.Get(url + "/engine-rest/deployment/count")
		if err != nil {
			t.Error(err)
			return
		}
		count := Count{}
		err = json.NewDecoder(resp.Body).Decode(&count)
		if err != nil {
			t.Error(err)
			return
		}
		if count.Count != expected {
			t.Error(count, expected)
			return
		}
	}
}

func testCreateVid(v *vid.Vid, vid string, pid string) func(t *testing.T) {
	return func(t *testing.T) {
		err := v.SaveVidRelation(vid, pid)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testCreateProcess(url string, pidRef *string) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := deployProcess(url, "test", bpmnExample, svgExample, "owner", "test")
		if err != nil {
			t.Fatal(err)
		}
		deploymentId, ok := result["id"].(string)
		if !ok {
			t.Fatal(err)
		}
		*pidRef = deploymentId
	}
}

func testCreateNormalProcess(url string, v *vid.Vid, vidStr string, pidRef *string) func(t *testing.T) {
	return func(t *testing.T) {
		testCreateProcess(url, pidRef)(t)
		testCreateVid(v, vidStr, *pidRef)(t)
	}
}

func deployProcess(shard string, name string, xml string, svg string, owner string, source string) (result map[string]interface{}, err error) {
	result = map[string]interface{}{}
	boundary := "---------------------------" + time.Now().String()
	b := strings.NewReader(buildPayLoad(name, xml, svg, boundary, owner, source))
	resp, err := http.Post(shard+"/engine-rest/deployment/create", "multipart/form-data; boundary="+boundary, b)
	if err != nil {
		log.Println("ERROR: request to processengine ", err)
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func buildPayLoad(name string, xml string, svg string, boundary string, owner string, deploymentSource string) string {
	segments := []string{}
	if deploymentSource == "" {
		deploymentSource = "sepl"
	}

	segments = append(segments, "Content-Disposition: form-data; name=\"data\"; "+"filename=\""+name+".bpmn\"\r\nContent-Type: text/xml\r\n\r\n"+xml+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"diagram\"; "+"filename=\""+name+".svg\"\r\nContent-Type: image/svg+xml\r\n\r\n"+svg+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-name\"\r\n\r\n"+name+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-source\"\r\n\r\n"+deploymentSource+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"tenant-id\"\r\n\r\n"+owner+"\r\n")

	return "--" + boundary + "\r\n" + strings.Join(segments, "--"+boundary+"\r\n") + "--" + boundary + "--\r\n"
}
