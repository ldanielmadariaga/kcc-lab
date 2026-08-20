// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/annotations"
	"golang.org/x/tools/imports"
	"k8s.io/klog/v2"
)

type generatorBase struct {
	outputBaseDir  string
	generatedFiles map[generatedFileKey]*generatedFile
	errors         []error

	// handWrittenCache memoizes scanHandWrittenTypes per source directory.
	handWrittenCache map[string]*handWrittenTypes
}

func (g *generatorBase) init(outputBaseDir string) {
	g.outputBaseDir = outputBaseDir
	g.generatedFiles = make(map[generatedFileKey]*generatedFile)
	g.handWrittenCache = make(map[string]*handWrittenTypes)
}

// handWrittenTypes records what a package already defines by hand: the type names it declares, and
// the proto messages those types claim through a +kcc:*proto annotation.
//
// The generator never overwrites a hand-written type. It skips any message such a type claims, and
// separately it has to decide, earlier in the run, whether that message gets an ObservedState
// struct of its own. If those two decisions read from different sources they could disagree, and
// we would emit a field typed *XObservedState with no such struct written anywhere. Both read from
// this one snapshot instead.
type handWrittenTypes struct {
	// typeNames holds every type name declared in a hand-written file.
	typeNames map[string]bool
	// protoTagged holds every proto message named by a +kcc:*proto annotation in a hand-written
	// file, whichever of the four annotations was used.
	protoTagged map[string]bool
}

// ownedByGenerator reports whether the message is entirely the generator's to shape: no
// hand-written type claims it by annotation, and none has taken either of the two names we would
// generate for it.
func (h *handWrittenTypes) ownedByGenerator(protoFullName string, goName string, observedStateGoName string) bool {
	if h == nil {
		return true
	}
	return !h.protoTagged[protoFullName] && !h.typeNames[goName] && !h.typeNames[observedStateGoName]
}

// scanHandWrittenTypes reads the non-generated Go files in srcDir and records what they define,
// caching the result per directory. A missing directory is not an error: a service generated for
// the first time does not have one yet.
func (g *generatorBase) scanHandWrittenTypes(srcDir string) (*handWrittenTypes, error) {
	if cached, ok := g.handWrittenCache[srcDir]; ok {
		return cached, nil
	}

	out := &handWrittenTypes{
		typeNames:   make(map[string]bool),
		protoTagged: make(map[string]bool),
	}

	files, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			g.handWrittenCache[srcDir] = out
			return out, nil
		}
		return nil, fmt.Errorf("reading directory %q: %w", srcDir, err)
	}

	for _, f := range files {
		p := filepath.Join(srcDir, f.Name())
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "generated.go") {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", p, err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if protoMessage, _, ok := GetProtoMessageAndKindFromAnnotation(line); ok {
				out.protoTagged[protoMessage] = true
				continue
			}
			if name, ok := goTypeNameFromDeclaration(line); ok {
				out.typeNames[name] = true
			}
		}
	}

	g.handWrittenCache[srcDir] = out
	return out, nil
}

// goTypeNameFromDeclaration pulls "Foo" out of a "type Foo struct {" line.
func goTypeNameFromDeclaration(line string) (string, bool) {
	rest, ok := strings.CutPrefix(line, "type ")
	if !ok {
		return "", false
	}
	name, _, ok := strings.Cut(rest, " ")
	if !ok || name == "" {
		return "", false
	}
	return name, true
}

func (g *generatorBase) getOutputFile(k generatedFileKey) *generatedFile {
	out := g.generatedFiles[k]
	if out == nil {
		out = &generatedFile{key: k, baseDir: g.outputBaseDir}
		g.generatedFiles[k] = out
	}
	return out
}

func (g *generatorBase) Errorf(msg string, args ...any) {
	g.errors = append(g.errors, fmt.Errorf(msg, args...))
}

type generatedFile struct {
	baseDir   string
	key       generatedFileKey
	goPackage string
	body      bytes.Buffer

	fileAnnotation *annotations.FileAnnotation

	imports map[string]string
}

type generatedFileKey struct {
	GoPackage string

	FileName string
}

func (f *generatedFile) OutputDir() string {
	tokens := strings.Split(f.key.GoPackage, ".")
	dirTokens := []string{f.baseDir}
	dirTokens = append(dirTokens, tokens...)
	dir := filepath.Join(dirTokens...)
	return dir
}

func (f *generatedFile) addImport(alias string, pkgName string) {
	if f.imports == nil {
		f.imports = make(map[string]string)
	}
	f.imports[pkgName] = alias
}

// usedImports drops imports the rendered body never references.
//
// The qualifier is the alias where one is set, otherwise the last path segment,
// which is how Go resolves it. Matching on "<qualifier>." is deliberately
// conservative: it can only ever keep an import that is genuinely unused if the
// string appears in a comment, and keeping one costs nothing next to emitting an
// import that breaks the build.
func (f *generatedFile) usedImports() map[string]string {
	if len(f.imports) == 0 {
		return nil
	}
	body := f.body.String()
	used := make(map[string]string, len(f.imports))
	for pkgName, alias := range f.imports {
		qualifier := alias
		if qualifier == "" {
			qualifier = lastGoComponent(pkgName)
		}
		if strings.Contains(body, qualifier+".") {
			used[pkgName] = alias
		}
	}
	return used
}

func (f *generatedFile) Write(addCopyright bool, writeEmptyFiles bool) error {
	if f.body.Len() == 0 && !writeEmptyFiles {
		return nil
	}

	dir := f.OutputDir()
	p := filepath.Join(dir, f.key.FileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}

	var w bytes.Buffer

	if addCopyright {
		year := time.Now().Year()

		b, err := os.ReadFile(p)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("reading file %q (checking for existing copyright): %w", p, err)
		} else {
			existingCopyrightYear, ok := extractCopyrightYear(b)
			if ok {
				year = existingCopyrightYear
			}
		}
		writeCopyright(&w, year)
	}

	fmt.Fprintf(&w, "// Code generated by dev/tasks/generate-all. DO NOT EDIT.\n")

	if f.key.FileName == "mapper.generated.go" {
		fmt.Fprintf(&w, "//go:build !ignore_autogenerated\n")
		fmt.Fprintf(&w, "// +build !ignore_autogenerated\n\n")
	}

	if f.fileAnnotation != nil {
		s := f.fileAnnotation.FormatGo()
		fmt.Fprintf(&w, "%s\n", s)
	}

	if f.goPackage != "" {
		fmt.Fprintf(&w, "package %s\n", f.goPackage)
		fmt.Fprintf(&w, "\n")
	}

	// Imports are added while walking a message's dependencies, before we know
	// whether that message will actually be written: a hand-written type may
	// claim it, or the scaffolder may take it. Either way the import is left
	// behind with nothing referencing it, which does not compile. Decide from
	// the rendered body instead of from the intent.
	used := f.usedImports()
	if len(used) != 0 {
		w.WriteString("import (\n")
		for pkgName, alias := range used {
			if alias == "" {
				w.WriteString(fmt.Sprintf("\t%q\n", pkgName))
			} else {
				w.WriteString(fmt.Sprintf("\t%s %q\n", alias, pkgName))
			}
		}
		w.WriteString(")\n")
	}
	f.body.WriteTo(&w)

	formatted, err := imports.Process(p, w.Bytes(), nil)
	if err != nil {
		klog.Errorf("error from goimports processing file %q:\n%s", p, w.String())
		return fmt.Errorf("running goimports on %q: %w", p, err)
	}

	klog.V(2).Infof("writing file %v", p)
	if err := os.WriteFile(p, formatted, 0644); err != nil {
		return fmt.Errorf("writing %q: %w", p, err)
	}

	return nil
}

// extractCopyrightYear extracts the copyright year from the given file bytes
func extractCopyrightYear(b []byte) (int, bool) {
	s := string(b)
	for _, line := range strings.Split(s, "\n") {
		if suffix, ok := strings.CutPrefix(line, "// Copyright "); ok {
			tokens := strings.Fields(suffix)
			if len(tokens) >= 1 {
				year, err := strconv.Atoi(tokens[0])
				if err == nil {
					return year, true
				}
			}
		}
	}

	return 0, false
}

func (g *generatorBase) WriteFiles(addCopyright bool, writeEmptyFiles bool) error {
	for _, f := range g.generatedFiles {
		if err := f.Write(addCopyright, writeEmptyFiles); err != nil {
			return err
		}
	}
	return nil
}

func (g *generatorBase) findTypeDeclaration(goTypeName string, srcDir string, skipGenerated bool) (*string, error) {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) { // type declaration does not exist
			return nil, nil
		}
		return nil, fmt.Errorf("reading directory %q: %w", srcDir, err)
	}

	for _, f := range files {
		p := filepath.Join(srcDir, f.Name())
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		if skipGenerated && strings.HasSuffix(p, "generated.go") {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", p, err)
		}

		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "type "+goTypeName+" ") {
				return &line, nil
			}
		}
	}

	return nil, nil
}

func (g *generatorBase) findTypeDeclarationWithProtoTag(protoTag string, srcDir string, skipGenerated bool) (*string, error) {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", srcDir, err)
	}

	for _, f := range files {
		p := filepath.Join(srcDir, f.Name())
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		if skipGenerated && strings.HasSuffix(p, "generated.go") {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", p, err)
		}

		for _, line := range strings.Split(string(b), "\n") {
			if proto, ok := GetProtoMessageFromAnnotation(line); ok {
				if proto == protoTag {
					return &line, nil
				}
			}
		}
	}

	return nil, nil
}

func (g *generatorBase) findFuncDeclaration(goFuncName string, srcDir string, skipGenerated bool) *string {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		g.Errorf("reading directory %q: %w", srcDir, err)
		return nil
	}

	for _, f := range files {
		p := filepath.Join(srcDir, f.Name())
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		if skipGenerated && strings.HasSuffix(p, "generated.go") {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			g.Errorf("reading file %q: %w", p, err)
			return nil
		}

		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "func "+goFuncName+"(") {
				return &line
			}
		}
	}

	return nil
}
