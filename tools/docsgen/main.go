package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Subcommands []Subcommand `yaml:"subcommands"`
	Common      Common       `yaml:"common"`
}

type Common struct {
	Flags []Flag `yaml:"flags"`
}

type Subcommand struct {
	ID          string    `yaml:"id"`
	Short       string    `yaml:"short"`
	Description string    `yaml:"description"`
	Usage       string    `yaml:"usage"`
	Flags       []Flag    `yaml:"flags"`
	Examples    []Example `yaml:"examples"`
	Notes       []string  `yaml:"notes,omitempty"`
}

type Flag struct {
	ID          string `yaml:"id"`
	Syntax      string `yaml:"syntax"`
	Description string `yaml:"description"`
	Default     string `yaml:"default,omitempty"`
	More        string `yaml:"more,omitempty"`
}

type Example struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

type TemplateData struct {
	Subcommand
	Date    string
	Version string
	IDUpper string
}

type Outputs struct {
	Template string
	Folder   string
	Prefix   string
	Suffix   string
}

func main() {

	docs := os.Args[1]

	data, _ := os.ReadFile(docs + "/templates/tfctl.yaml")
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(err)
	}

	for _, sub := range config.Subcommands {
		mergedFlags := config.Common.Flags

		if len(sub.Flags) > 0 {
			mergedFlags = append(mergedFlags, sub.Flags...)
		}

		sort.Slice(mergedFlags, func(i, j int) bool {
			return mergedFlags[i].ID < mergedFlags[j].ID
		})
		sub.Flags = mergedFlags

		// Prepare template data
		metadata := TemplateData{
			Subcommand: sub,
			Date:       time.Now().Format("January 2, 2006"),
			Version:    getVersion(),
		}

		types := []Outputs{
			{Template: docs + "/templates/tfctl.md.tmpl", Folder: docs + "/commands/", Suffix: ".md"},
			{Template: docs + "/templates/tfctl.man.tmpl", Folder: docs + "/./man/share/man1/", Prefix: "tfctl-", Suffix: ".1"},
			{Template: docs + "/templates/tfctl.tldr.tmpl", Folder: docs + "/./tldr/", Prefix: "tfctl-", Suffix: ".md"},
		}

		for _, t := range types {
			if err := os.MkdirAll(t.Folder, 0755); err != nil {
				panic(err)
			}

			file, _ := os.Create(t.Folder + t.Prefix + sub.ID + t.Suffix)
			fmt.Println("Generating", t.Folder+t.Prefix+sub.ID+t.Suffix)
			tmpl, err := template.ParseFiles(t.Template)
			if err != nil {
				panic(err)
			}

			if err := tmpl.Execute(file, metadata); err != nil {
				panic(err)
			}

			file.Close()
		}
	}
}

// getVersion returns the version string from git tags, stripping the leading
// "v" prefix. Falls back to "dev" if git describe fails.
func getVersion() string {
	out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return "dev"
	}

	version := strings.TrimSpace(string(out))
	return strings.TrimPrefix(version, "v")
}
