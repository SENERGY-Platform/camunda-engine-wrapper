package events

const CurrentVersion int64 = 3

type Deployment struct {
	Version int64   `json:"version"`
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Diagram Diagram `json:"diagram"`
}

type Diagram struct {
	XmlDeployed string `json:"xml_deployed"`
	Svg         string `json:"svg"`
}

type DeploymentCommand struct {
	Command    string      `json:"command"`
	Id         string      `json:"id"`
	Owner      string      `json:"owner"`
	Deployment *Deployment `json:"deployment"`
	Source     string      `json:"source,omitempty"`
	Version    int64       `json:"version"`
}

type VersionWrapper struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Version int64  `json:"version"`
	Owner   string `json:"owner"`
}

type KafkaIncidentsCommand struct {
	Command             string `json:"command"`
	MsgVersion          int64  `json:"msg_version"`
	ProcessDefinitionId string `json:"process_definition_id,omitempty"`
	ProcessInstanceId   string `json:"process_instance_id,omitempty"`
}

type DoneNotification struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Handler string `json:"handler"`
}
