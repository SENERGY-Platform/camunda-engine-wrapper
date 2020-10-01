package events

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/etree"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/camunda"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/configuration"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/events/kafka"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/lib/vid"
	"log"
)

type Events struct {
	zookeeperUrl    string
	kafkaGroupId    string
	deploymentTopic string
	incidentsTopic  string
	debug           bool
	vid             *vid.Vid
	camunda         *camunda.Camunda
	cqrs            kafka.Interface
}

func New(config configuration.Config, cqrs kafka.Interface, vid *vid.Vid, camunda *camunda.Camunda) (events *Events, err error) {
	events = &Events{
		zookeeperUrl:    config.ZookeeperUrl,
		kafkaGroupId:    config.KafkaGroup,
		deploymentTopic: config.DeploymentTopic,
		incidentsTopic:  config.IncidentTopic,
		debug:           config.Debug,
		vid:             vid,
		camunda:         camunda,
		cqrs:            cqrs,
	}
	err = events.init()
	return
}

func (this *Events) init() (err error) {
	err = this.cqrs.Consume(this.deploymentTopic, func(delivery []byte) error {
		command := DeploymentCommand{}
		err = json.Unmarshal(delivery, &command)
		if err != nil {
			log.Println("ERROR: unable to parse cqrs event as json \n", err, "\n --> ignore event \n", string(delivery))
			return nil
		}
		log.Println("cqrs receive ", string(delivery))
		switch command.Command {
		case "PUT":
			owner, id, name, xml, svg, source, err := parsePutCommand(command)
			if err != nil {
				return err
			}
			return this.HandleDeploymentCreate(owner, id, name, xml, svg, source)
		case "POST":
			log.Println("WARNING: deprecated event type POST")
			return nil
		case "DELETE":
			return this.HandleDeploymentDelete(command.Id, command.Owner)
		default:
			log.Println("WARNING: unknown event type", string(delivery))
			return nil
		}
	})
	return err
}

func parsePutCommand(command DeploymentCommand) (owner string, id string, name string, xml string, svg string, source string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered error %v", r)
		}
	}()
	if command.Deployment != nil {
		return parseV1Command(command)
	} else if command.DeploymentV2 != nil {
		return parseV2Command(command)
	}
	err = errors.New("unknown version")
	return
}

func parseV1Command(command DeploymentCommand) (owner string, id string, name string, xml string, svg string, source string, err error) {
	return command.Owner, command.Id, command.Deployment.Name, command.Deployment.Xml, command.Deployment.Svg, command.Source, nil
}

func parseV2Command(command DeploymentCommand) (owner string, id string, name string, xml string, svg string, source string, err error) {
	return command.Owner, command.Id, command.DeploymentV2.Name, command.DeploymentV2.Diagram.XmlDeployed, command.DeploymentV2.Diagram.Svg, command.Source, err
}

func (this *Events) HandleDeploymentDelete(vid string, userId string) error {
	id, exists, err := this.vid.GetDeploymentId(vid)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = this.deleteIncidentsByDeploymentId(id, userId)
	if err != nil {
		return err
	}

	commit, rollback, err := this.vid.RemoveVidRelation(vid, id)
	if err != nil {
		return err
	}
	if userId != "" {
		err = this.camunda.RemoveProcess(id, userId)
	} else {
		err = this.camunda.RemoveProcessFromAllShards(id)
	}
	if err != nil {
		rollback()
	} else {
		commit()
	}
	return err
}

func (this *Events) HandleDeploymentCreate(owner string, id string, name string, xml string, svg string, source string) (err error) {
	err = this.cleanupExistingDeployment(id, owner)
	if err != nil {
		return err
	}
	if !validateXml(xml) {
		log.Println("ERROR: got invalid xml, replace with default")
		xml = camunda.CreateBlankProcess()
		svg = camunda.CreateBlankSvg()
	}
	if this.debug {
		log.Println("deploy process", id, name, xml)
	}
	deploymentId, err := this.camunda.DeployProcess(name, xml, svg, owner, source)
	if err != nil {
		log.Println("WARNING: unable to deploy process to camunda ", err)
		return err
	}
	if this.debug {
		log.Println("save vid relation", id, deploymentId)
	}
	err = this.vid.SaveVidRelation(id, deploymentId)
	if err != nil {
		log.Println("WARNING: unable to publish deployment saga \n", err, "\nremove deployed process")
		removeErr := this.camunda.RemoveProcess(deploymentId, owner)
		if removeErr != nil {
			log.Println("ERROR: unable to remove deployed process", deploymentId, err)
		}
		return err
	}
	return err
}

func validateXml(xmlStr string) bool {
	if xmlStr == "" {
		return false
	}
	err := etree.NewDocument().ReadFromString(xmlStr)
	if err != nil {
		log.Println("ERROR: unable to parse xml", err)
		return false
	}
	err = xml.Unmarshal([]byte(xmlStr), new(interface{}))
	if err != nil {
		log.Println("ERROR: unable to parse xml", err)
		return false
	}
	return true
}

func (this *Events) cleanupExistingDeployment(vid string, userId string) error {
	exists, err := this.vid.VidExists(vid)
	if err != nil {
		return err
	}
	if exists {
		return this.HandleDeploymentDelete(vid, userId)
	}
	return nil
}

func (this *Events) deleteIncidentsByDeploymentId(id string, userId string) (err error) {
	definitions, err := this.camunda.GetRawDefinitionsByDeployment(id, userId)
	if err != nil {
		return err
	}
	for _, definition := range definitions {
		err = this.PublishIncidentsDeleteByProcessDefinitionEvent(definition.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Events) PublishIncidentsDeleteByProcessDefinitionEvent(definitionId string) error {
	command := KafkaIncidentsCommand{
		Command:             "DELETE",
		ProcessDefinitionId: definitionId,
		MsgVersion:          3,
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return this.cqrs.Publish(this.incidentsTopic, definitionId, payload)
}

func (this *Events) PublishIncidentDeleteByProcessInstanceEvent(instanceId string, definitionId string) error {
	command := KafkaIncidentsCommand{
		Command:           "DELETE",
		ProcessInstanceId: instanceId,
		MsgVersion:        3,
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}
	return this.cqrs.Publish(this.incidentsTopic, definitionId, payload)
}
