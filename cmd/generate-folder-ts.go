package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type layer struct {
	name       string
	subFolders []string
	extensions []extension
}

type extension struct {
	name         string
	insideFolder bool
	folderName   string
}

var (
	layers = [3]layer{
		{
			name:       "presentation",
			subFolders: []string{},
			extensions: []extension{{name: "controller"}, {name: "request"}, {name: "response"}},
		},
		{
			name:       "domain",
			subFolders: []string{"objects"},
			extensions: []extension{
				{name: "input", insideFolder: true, folderName: "objects"},
				{name: "interactor"},
				{name: "repository"},
			},
		},
		{
			name:       "infrastructure",
			subFolders: []string{},
			extensions: []extension{{name: "postgres-repository"}},
		},
	}
	//go:embed folder-ts-templates/*.txt
	templatesFS embed.FS
)

var generateFolderTSCmd = &cobra.Command{
	Use:   "generate-folder-ts",
	Short: "generate clean-architecture folder-for-feature (TS)",

	Run: generateFolderTS,
}

func init() {
	rootCmd.AddCommand(generateFolderTSCmd)

	generateFolderTSCmd.Flags().StringP("feature-name", "n", "", "name of the feature")
	generateFolderTSCmd.MarkFlagRequired("feature-name")
}

func generateFolderTS(cmd *cobra.Command, args []string) {
	featureName, err := cmd.Flags().GetString("feature-name")
	if err != nil {
		log.Fatalf("error parsing -n flag: %s", err)
	}

	if err := createFolders(featureName); err != nil {
		log.Fatalf("error creating folders: %s", err)
	}

	if err := createFiles(featureName); err != nil {
		log.Fatalf("error creating files: %s", err)
	}
}

func createFolders(featureName string) error {
	if err := os.Mkdir(featureName, 0755); err != nil {
		return err
	}

	for _, layer := range layers {
		layerFolderPath := fmt.Sprintf("%s/%s", featureName, layer.name)

		if err := os.Mkdir(layerFolderPath, 0755); err != nil {
			return err
		}

		if len(layer.subFolders) != 0 {
			for _, subFolder := range layer.subFolders {
				subFolderPath := fmt.Sprintf("%s/%s", layerFolderPath, subFolder)

				if err := os.Mkdir(subFolderPath, 0755); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func createFiles(featureName string) error {
	if err := createAndWriteFileFromTemplate(featureName, "", "module"); err != nil {
		return err
	}

	for _, layer := range layers {
		for _, extension := range layer.extensions {
			layerName := layer.name

			if extension.insideFolder {
				layerName = fmt.Sprintf("%s/%s", layer.name, extension.folderName)
			}

			err := createAndWriteFileFromTemplate(featureName, layerName, extension.name)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func createAndWriteFileFromTemplate(featureName, layer, extension string) error {
	templatePath := fmt.Sprintf("folder-ts-templates/%s.txt", extension)

	templateBytes, err := fs.ReadFile(templatesFS, templatePath)
	if err != nil {
		return err
	}

	templateCode := strings.ReplaceAll(string(templateBytes), "{{feature_name}}", featureName)
	templateCode = strings.ReplaceAll(templateCode, "{{feature_name_camel_case}}", toCamelCase(featureName))

	filePath := fmt.Sprintf("%s/%s/%s.%s.ts", featureName, layer, featureName, extension)

	if layer == "" {
		filePath = fmt.Sprintf("%s/%s.%s.ts", featureName, featureName, extension)
	}

	if err := os.WriteFile(filePath, []byte(templateCode), 0644); err != nil {
		return err
	}

	return nil
}

func toCamelCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)

	for i, word := range words {
		words[i] = cases.Title(language.English, cases.NoLower).String(word)
	}

	return strings.Join(words, "")
}
