package helper

import (
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/auth"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/client"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/model"
	"net/http"
)

func PutProcess(c *client.Client, vid string, name string, owner string) error {
	err, _ := c.Deploy(client.InternalAdminToken, client.DeploymentMessage{
		Deployment: model.Deployment{
			Id:   vid,
			Name: name,
			Diagram: model.Diagram{
				XmlDeployed: BpmnExample,
				Svg:         SvgExample,
			},
			IncidentHandling: &model.IncidentHandling{},
		},
		UserId: owner,
	})
	return err
}

func PutProcessWithSource(c *client.Client, vid string, name string, owner string, source string) error {
	err, _ := c.Deploy(client.InternalAdminToken, client.DeploymentMessage{
		Deployment: model.Deployment{
			Id:   vid,
			Name: name,
			Diagram: model.Diagram{
				XmlDeployed: BpmnExample,
				Svg:         SvgExample,
			},
			IncidentHandling: &model.IncidentHandling{},
		},
		UserId: owner,
		Source: source,
	})
	return err
}

func DeleteProcess(c *client.Client, vid string, userId string) error {
	err, _ := c.DeleteDeployment(client.InternalAdminToken, userId, vid)
	return err
}

const BpmnExample = `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' xmlns:di='http://www.omg.org/spec/DD/20100524/DI' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='Process_1' isExecutable='true'><bpmn:startEvent id='StartEvent_1'><bpmn:outgoing>SequenceFlow_02ibfc0</bpmn:outgoing></bpmn:startEvent><bpmn:endEvent id='EndEvent_1728wjv'><bpmn:incoming>SequenceFlow_02ibfc0</bpmn:incoming></bpmn:endEvent><bpmn:sequenceFlow id='SequenceFlow_02ibfc0' sourceRef='StartEvent_1' targetRef='EndEvent_1728wjv'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='Process_1'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNShape id='EndEvent_1728wjv_di' bpmnElement='EndEvent_1728wjv'><dc:Bounds x='259' y='102' width='36' height='36'/></bpmndi:BPMNShape><bpmndi:BPMNEdge id='SequenceFlow_02ibfc0_di' bpmnElement='SequenceFlow_02ibfc0'><di:waypoint x='209' y='120'></di:waypoint><di:waypoint x='259' y='120'></di:waypoint></bpmndi:BPMNEdge></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
const SvgExample = `<svg height='48' version='1.1' viewBox='167 96 134 48' width='134' xmlns='http://www.w3.org/2000/svg' xmlns:xlink='http://www.w3.org/1999/xlink'><defs><marker id='sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p' markerHeight='10' markerWidth='10' orient='auto' refX='11' refY='10' viewBox='0 0 20 20'><path d='M 1 5 L 11 10 L 1 15 Z' style='fill: black; stroke-width: 1px; stroke-linecap: round; stroke-dasharray: 10000, 1; stroke: black;'/></marker></defs><g class='djs-group'><g class='djs-element djs-connection' data-element-id='SequenceFlow_04zz9eb' style='display: block;'><g class='djs-visual'><path d='m  209,120L259,120 ' style='fill: none; stroke-width: 2px; stroke: black; stroke-linejoin: round; marker-end: url(&apos;#sequenceflow-end-white-black-3gh21e50i1p8scvqmvrotmp9p&apos;);'/></g><polyline class='djs-hit' points='209,120 259,120 ' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;'/><rect class='djs-outline' height='12' style='fill: none;' width='62' x='203' y='114'/></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='StartEvent_1' style='display: block;' transform='translate(173 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 2px; fill: white; fill-opacity: 0.95;'/><path d='m 8.459999999999999,11.34 l 0,12.6 l 18.900000000000002,0 l 0,-12.6 z l 9.450000000000001,5.4 l 9.450000000000001,-5.4' style='fill: white; stroke-width: 1px; stroke: black;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g><g class='djs-group'><g class='djs-element djs-shape' data-element-id='EndEvent_056p30q' style='display: block;' transform='translate(259 102)'><g class='djs-visual'><circle cx='18' cy='18' r='18' style='stroke: black; stroke-width: 4px; fill: white; fill-opacity: 0.95;'/></g><rect class='djs-hit' height='36' style='fill: none; stroke-opacity: 0; stroke: white; stroke-width: 15px;' width='36' x='0' y='0'></rect><rect class='djs-outline' height='48' style='fill: none;' width='48' x='-6' y='-6'></rect></g></g></svg>`

const Jwt string = `Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJsWVh1Y1NFMHVQcFpDUHhZX3Q1WEVnMlRsWUoyTVl0TWhwN1hLNThsbmJvIn0.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiIzN2MyM2QzMC00YjQ4LTQyMDktOWJkNy0wMzcxZjYyYzJjZmYiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXSwicHJlZmVycmVkX3VzZXJuYW1lIjoic2VwbCJ9.cSWTHIOHkugQcVNgatbXjvDIP_Ir_QKuUuozbyweh1dJEFsZToTjJ4-5w947bLETmqiNElqXlIV8dT4c9DnPoiXAzsdSotkzKFEYEqRhjYm2obc7Wine1rVwFC4b0Tc5voIzCPNVGFlJDFYWqsPuQYNvAuCIs_A4W86AXWAuxzTyBk5gcRVBLLkFX6GErS2a_4jKd0m26Wd3qoO_j5cl2z2r0AtJ5py4PESiTRLDxEiMoahVQ4coYtX2esWoCRpkSa-beqlD8ffuKaHt95Z8AVcGjBZeSuZpVq6qY6bPBasqVdNkq-CvSnXqWnzNhvq2lUPt58Wp7jeMIJQG4015Zg`

var JwtPayload, _ = auth.Parse(Jwt)

func JwtGet(token string, endpoint string) (resp *http.Response, err error) {
	req, err := http.NewRequest(
		"GET",
		endpoint,
		nil,
	)
	if err != nil {
		return resp, err
	}
	req.Header.Set("Authorization", token)
	return http.DefaultClient.Do(req)
}
