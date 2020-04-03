// +build ignore

/*
Copyright 2020 Intel Corporation

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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const (
	blocksPerRow = 12
)

type Program struct {
	ProgramLines []string
}

func main() {
	f, err := ioutil.ReadFile("../../libexec/avx512.o")
	if err != nil {
		fmt.Println("Note: AVX512 eBPF ELF not available.")
	}
	enc := make([]byte, hex.EncodedLen(len(f)))
	enclen := hex.Encode(enc, f)

	var j int
	var row strings.Builder
	program := make([]string, 0)

	for i := 0; i < enclen-1; i = i + 2 {
		fmt.Fprintf(&row, "0x%s, ", enc[i:i+2])
		j++
		if j%blocksPerRow == 0 {
			program = append(program, row.String())
			row.Reset()
		}
	}
	// flush last row
	program = append(program, row.String())

	p := Program{
		ProgramLines: program,
	}

	template := template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.

package avx

var program = [...]byte{
{{- range .ProgramLines }}
	{{ printf "%s" . }}
{{- end }}
}
`))

	outfile, err := os.Create("programbytes_gendata.go")
	if err != nil {
		fmt.Println("elfdump:", err)
		os.Exit(1)
	}
	defer outfile.Close()

	err = template.Execute(outfile, p)
	if err != nil {
		fmt.Println("elfdump:", err)
	}
}
