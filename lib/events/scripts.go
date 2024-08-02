/*
 * Copyright 2024 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package events

import (
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/camunda-engine-wrapper/etree"
	"log"
	"runtime/debug"
)

func SecureProcessScripts(xml string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			log.Printf("%s: %s", r, debug.Stack())
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
		}
	}()
	doc := etree.NewDocument()
	err = doc.ReadFromString(xml)
	if err != nil {
		return result, err
	}
	prefix := `var java = {}; var execution = {}; `

	for _, script := range doc.FindElements("//camunda:script") {
		script.SetText(prefix + script.Text())
	}
	for _, script := range doc.FindElements("//bpmn:script") {
		script.SetText(prefix + script.Text())
	}

	return doc.WriteToString()
}
