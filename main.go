package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var gofileTpl = `package main
{{range $import := .imports}}
{{$import}}
{{end}}
func main() {
	{{.body}}
}`

func run(args []string) error {
	var gomakefile string
	{
		fbytes, err := os.ReadFile("./Gomakefile")
		if err != nil {
			return err
		}
		gomakefile = string(fbytes)
	}

	regexpGoStatement := regexp.MustCompile(`(?s)go \{(?P<body>.*)}`)
	bodyIndex := regexpGoStatement.SubexpIndex("body")
	goStatements := regexpGoStatement.FindAllStringSubmatch(gomakefile, -1)

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("gomakebuild%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}

	for _, statement := range goStatements {
		body := statement[bodyIndex]
		regexpGoImport := regexp.MustCompile(`(?s)import "\S+"`)
		goImports := regexpGoImport.FindAllString(body, -1)

		body = regexpGoImport.ReplaceAllString(body, "")

		tplEngine, err := template.New("gofile").Parse(gofileTpl)
		if err != nil {
			return err
		}

		entryPoint := fmt.Sprintf("%s/gobuild%d/main.go", tempDir, time.Now().UnixNano())
		if err = os.Mkdir(filepath.Dir(entryPoint), os.ModePerm); err != nil {
			return err
		}
		source, err := os.Create(entryPoint)
		if err != nil {
			return err
		}

		err = tplEngine.Execute(source, map[string]any{
			"imports": goImports,
			"body":    body,
		})
		source.Close()
		if err != nil {
			return err
		}

		executeGoStmt := fmt.Sprintf("execgo%d", time.Now().UnixNano())
		cmd := fmt.Sprintf(`cd %s; go mod init gomakegen; go mod tidy; go install golang.org/x/tools/cmd/goimports@latest; goimports -w main.go; go build -o %s/%s; cd %s`,
			filepath.Dir(entryPoint), tempDir, executeGoStmt, tempDir)
		cmdExecutor := exec.Command("bash", "-c", cmd)
		cmdExecutor.Stdout = os.Stdout
		cmdExecutor.Stderr = os.Stderr

		if err = cmdExecutor.Run(); err != nil {
			return err
		}

		nextGoStatement := regexpGoStatement.FindString(gomakefile)
		ahead := strings.Split(gomakefile, nextGoStatement)[0]
		if ahead[len(ahead)-1] == '=' {
			executeGoStmt = fmt.Sprintf("$(%s/%s)", tempDir, executeGoStmt)
		} else {
			executeGoStmt = fmt.Sprintf("%s/%s", tempDir, executeGoStmt)
		}

		gomakefile = strings.Replace(gomakefile, nextGoStatement, executeGoStmt, 1)
	}

	regexpGlobalScope := regexp.MustCompile(`(?Us)(?P<block>.*)(\S+:)`)
	globalScopeBlockIndex := regexpGlobalScope.SubexpIndex("block")
	globalScope := regexpGlobalScope.FindStringSubmatch(gomakefile)[globalScopeBlockIndex]
	gomakefile = strings.ReplaceAll(gomakefile, globalScope, "")

	regexpTargetStatement := regexp.MustCompile(`(?Um)(?P<name>\S+):([\w ]+)?\n(?P<block>(.*\n)+)((\S+:([\w ]+)?)|$)`)
	targetBlockIndex, targetNameIndex :=
		regexpTargetStatement.SubexpIndex("block"),
		regexpTargetStatement.SubexpIndex("name")
	targets := regexpTargetStatement.FindAllStringSubmatch(gomakefile, -1)
	for _, target := range targets {
		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		name := target[targetNameIndex] + "target"
		oldBlock := target[targetBlockIndex]
		newBlock := fmt.Sprintf("cd %s\n%s\n%s", workingDir, globalScope, oldBlock)
		if err = os.WriteFile(fmt.Sprintf("%s/%s", tempDir, name), []byte(newBlock), os.ModePerm); err != nil {
			return err
		}
		gomakefile = strings.ReplaceAll(gomakefile, oldBlock, fmt.Sprintf("\t./%s\n", name))
	}

	gomakefile = strings.ReplaceAll(gomakefile, "    ", "\t")
	err = os.WriteFile(fmt.Sprintf("%s/Makefile", tempDir), []byte(gomakefile), os.ModePerm)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("cd %s; make %s", tempDir, strings.Join(args, ""))
	cmdExecutor := exec.Command("bash", "-c", cmd)
	cmdExecutor.Stdout = os.Stdout
	cmdExecutor.Stdin = os.Stdin
	cmdExecutor.Stderr = os.Stderr
	return cmdExecutor.Run()
}

func main() {
	args := os.Args[1:]
	if err := run(args); err != nil {
		panic(err)
	}
}
