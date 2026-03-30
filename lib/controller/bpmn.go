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

package controller

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/camunda-engine-wrapper/etree"
)

func SecureProcessScripts(xml string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
			slog.Error("recover from panic in SecureProcessScripts", "error", err, "stack", string(debug.Stack()))
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

func SetProcessId(xml string, id string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil && err == nil {
			err = errors.New(fmt.Sprint("Recovered Error: ", r))
			slog.Error("recover from panic in SetProcessId", "error", err, "stack", string(debug.Stack()))
		}
	}()
	doc := etree.NewDocument()
	err = doc.ReadFromString(xml)
	if err != nil {
		return result, err
	}
	normalizedId := "deplid_" + strings.NewReplacer("-", "_", ":", "_", "#", "_").Replace(id)
	for i, element := range doc.FindElements("//bpmn:process") {
		attr := element.SelectAttr("id")
		if attr != nil {
			if i > 0 {
				attr.Value = normalizedId + "_" + strconv.Itoa(i)
			} else {
				attr.Value = normalizedId
			}
		}
	}
	return doc.WriteToString()
}
