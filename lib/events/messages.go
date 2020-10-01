package events

type DeploymentV1 struct {
	Id   string `json:"id"`
	Xml  string `json:"xml"`
	Svg  string `json:"svg"`
	Name string `json:"name"`
}

type DeploymentV2 struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Diagram Diagram `json:"diagram"`
}

type Diagram struct {
	XmlDeployed string `json:"xml_deployed"`
	Svg         string `json:"svg"`
}

type DeploymentCommand struct {
	Command      string        `json:"command"`
	Id           string        `json:"id"`
	Owner        string        `json:"owner"`
	Deployment   *DeploymentV1 `json:"deployment"`
	DeploymentV2 *DeploymentV2 `json:"deployment_v2"`
	Source       string        `json:"source,omitempty"`
}

type KafkaIncidentsCommand struct {
	Command             string `json:"command"`
	MsgVersion          int64  `json:"msg_version"`
	ProcessDefinitionId string `json:"process_definition_id,omitempty"`
	ProcessInstanceId   string `json:"process_instance_id,omitempty"`
}
