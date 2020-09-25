package vidcleanup

import (
	"encoding/json"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/kafka"
)

func removeVidByEvent(cqrs kafka.Interface, topic string, vid string) error {
	command := DeploymentDeleteCommand{Id: vid, Command: "DELETE"}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return cqrs.Publish(topic, vid, payload)
}

type DeploymentDeleteCommand struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Owner   string `json:"owner"`
	Source  string `json:"source,omitempty"`
}
