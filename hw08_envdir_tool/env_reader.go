package main

import (
	"bytes"
	"io"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	env := make(Environment)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, "=") {
			continue
		}
		file, err := os.Open(dir + "/" + name)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		file.Close()
		lines := bytes.Split(data, []byte("\n"))
		var firstLine []byte
		if len(lines) > 0 {
			firstLine = lines[0]
		}
		value := bytes.ReplaceAll(firstLine, []byte{0x00}, []byte("\n"))
		strValue := strings.TrimRight(string(value), " \t")
		if len(strValue) == 0 {
			env[name] = EnvValue{NeedRemove: true}
		} else {
			env[name] = EnvValue{Value: strValue}
		}
	}
	return env, nil
}
