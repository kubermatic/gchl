/*
Copyright 2022 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	fields   = 100
	filename = "pkg/github/pullrequests_gen.go"
)

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func main() {
	templates, err := filepath.Glob("hack/*.go.tmpl")
	if err != nil {
		log.Fatalf("Failed to find Go templates: %v", err)
	}

	data := map[string]interface{}{
		"numFields": fields,
		"fields":    makeRange(0, fields-1),
	}

	for _, templateFile := range templates {
		log.Printf("Rendering %s...", templateFile)

		content, err := os.ReadFile(templateFile)
		if err != nil {
			log.Fatalf("Failed to read %s -- did you run this from the root directory?: %v", templateFile, err)
		}

		tpl := template.Must(template.New("tpl").Parse(string(content)))

		var buf bytes.Buffer
		tpl.Execute(&buf, data)

		source, err := format.Source(buf.Bytes())
		if err != nil {
			log.Fatalf("Failed to format generated code: %v", err)
		}

		filename := filepath.Join("pkg/github", strings.TrimSuffix(filepath.Base(templateFile), ".tmpl"))

		err = os.WriteFile(filename, source, 0644)
		if err != nil {
			log.Fatalf("Failed to write %s: %v", filename, err)
		}
	}
}
