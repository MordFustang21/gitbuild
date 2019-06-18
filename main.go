package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {

	// do git diff and get all files
	if len(os.Args) < 2 {
		log.Fatal("Directory is required")
	}

	dir := os.Args[1]

	files := getFiles(dir)

	packages := getPackageNames(dir, files)

	mains := findAllMains(os.Args[2])

	for main, deps := range mains {
		rebuild := false
		for _, dep := range deps {
			if contains(packages, dep) {
				rebuild = true
				break
			}
		}

		if rebuild {
			fmt.Println("Rebuilding", main)
		}
	}
}

func contains(list []string, key string) bool {
	for _, item := range list {
		if item == key {
			return true
		}
	}

	return false
}

// getFiles will find all changed files between current branch and master for now
func getFiles(dir string) []string {
	var modFiles []string

	// git diff command
	//branch := "master"
	cmd := exec.Command("git", "diff", "--name-only")

	// set working dir
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := bytes.Split(out, []byte{'\n'})
	for _, line := range lines {
		modFiles = append(modFiles, string(line))
	}

	return modFiles
}

// getPackageNames will grab the names of all packages that have been changed
func getPackageNames(dir string, files []string) []string {
	packages := make(map[string]struct{})
	for _, file := range files {
		path := filepath.Dir(file)
		dirPath := filepath.Join(dir, path)
		out, err := exec.Command("go", "list", "-json", dirPath).Output()
		if err != nil {
			continue
		}

		var list ListOut
		err = json.Unmarshal(out, &list)
		if err != nil {
			continue
		}

		packages[list.ImportPath] = struct{}{}
	}

	// convert map to array
	var packageList []string
	for key := range packages {
		packageList = append(packageList, key)
	}

	return packageList
}


type ListOut struct {
	Dir         string   `json:"Dir"`
	ImportPath  string   `json:"ImportPath"`
	Name        string   `json:"Name"`
	Target      string   `json:"Target"`
	Root        string   `json:"Root"`
	Match       []string `json:"Match"`
	Stale       bool     `json:"Stale"`
	StaleReason string   `json:"StaleReason"`
	GoFiles     []string `json:"GoFiles"`
	Imports     []string `json:"Imports"`
	Deps        []string `json:"Deps"`
}

// findAllMains will find all main packages and their dependent packages
// and returns a map with dir and all deps
func findAllMains(dir string) map[string][]string {
	mainDeps := make(map[string][]string)
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// check package
			cmd := exec.Command("go", "list", "-f", "{{.Name}}")

			cmd.Dir = path

			out, err := cmd.Output()
			if err != nil {
				return nil
			}

			// main package found get all deps
			if string(out) == "main\n" {
				cmd := exec.Command("go", "list", "-json", "-f", "{{.Deps}}")
				cmd.Dir = path

				out, err := cmd.Output()
				if err != nil {

				}

				var list ListOut
				err = json.Unmarshal(out, &list)
				if err != nil {

				}

				mainDeps[list.ImportPath] = list.Deps
			}
		}

		return nil
	})

	return mainDeps
}