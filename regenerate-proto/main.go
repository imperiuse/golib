package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type (
	protoFolder = string
	protoFile   = string

	ProtoDeclaration struct {
		PackageName string
		Folder      string
		Files       []string
	}
)

func main() {
	if err := app(); err != nil {
		log.Fatal("regenerate-proto failed with error: ", err)
	}
}

func app() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	protoMap, err := getProtoFiles(wd)
	if err != nil {
		return err
	}

	var pds []ProtoDeclaration
	for _, protoFiles := range protoMap {
		pd, err := NewProtoDeclaration(protoFiles)
		if err != nil {
			return err
		}
		pds = append(pds, pd)
	}

	// Create folders
	packageFolder := filepath.Join("internal", "grpc")
	if err := os.RemoveAll(packageFolder); err != nil {
		return err
	}

	if err := os.MkdirAll(packageFolder, os.ModePerm); err != nil {
		return err
	}

	for _, pd := range pds {
		protocString := "--go_out=import_path=" + pd.PackageName + ","
		for _, filename := range pd.Files {
			protocString += "M" + filename + "=" + pd.PackageName + ","
		}

		cmd := exec.Command("protoc",
			append([]string{"-I/usr/local/include", "-I" + pd.Folder, protocString + "plugins=grpc:" + packageFolder}, pd.Files...)...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err = cmd.Run()

		if err != nil {
			return err
		}
	}

	return nil
}

func NewProtoDeclaration(files []string) (ProtoDeclaration, error) {
	if len(files) == 0 {
		return ProtoDeclaration{}, io.ErrUnexpectedEOF
	}

	packageName, folder := getPackageNameAndFolder(files[0])
	return ProtoDeclaration{
		PackageName: packageName,
		Folder:      folder,
		Files:       files,
	}, nil
}

func getPackageNameAndFolder(filename string) (string, string) {
	path := filepath.Dir(filename)
	pathParts := strings.Split(path, string(filepath.Separator))
	for i := len(pathParts) - 1; i != 0; i-- {
		if i+1 < len(pathParts) && pathParts[i+1] == "proto" {
			return strings.Join(pathParts[i:], "_"), strings.Join(pathParts[:i], "/")
		}
	}

	panic("failed to get top level directory for proto files")
}

func getProtoFiles(dir string) (map[protoFolder][]protoFile, error) {
	protoPaths := map[protoFolder][]protoFile{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}

			matches, err := filepath.Glob(filepath.Join(path, "*.proto"))
			if err != nil {
				return err
			}

			if len(matches) != 0 {
				protoPaths[path] = matches
			}
		}
		return err
	})
	return protoPaths, err
}
