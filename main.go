package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"go/format"

	"github.com/FS-Frost/openapi2go/openapigen"
	"github.com/ettle/strcase"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("USAGE:\nopenapi2go spec1.yml [specN.yml] outDir")
		os.Exit(1)
	}

	outDir := os.Args[len(os.Args)-1]

	err := os.RemoveAll(outDir)
	checkError(err, "failed to delete directory %q", outDir)

	err = os.MkdirAll(outDir, os.ModePerm)
	checkError(err, "failed to create directory %q", outDir)

	for i := 1; i < len(os.Args); i++ {
		sourcePath := os.Args[i]
		if path.Ext(sourcePath) != ".yaml" && path.Ext(sourcePath) != ".json" {
			continue
		}

		fmt.Printf("source: %s\n", sourcePath)
		doc, err := openapi3.NewLoader().LoadFromFile(sourcePath)
		checkError(err, "failed to load spec file %q", sourcePath)

		saveUtils(outDir)

		fmt.Printf("paths: %d\n", doc.Paths.Len())
		for _, pathKey := range doc.Paths.InMatchingOrder() {
			fmt.Printf("path: %s\n", pathKey)
			docPath := doc.Paths.Find(pathKey)

			for operationName, operation := range docPath.Operations() {
				prefix := strcase.ToCamel(doc.Info.Title)
				prefix += pathKey
				prefix = strings.ReplaceAll(prefix, "/", "_")
				prefix = fmt.Sprintf("%s_%s", prefix, operationName)
				prefix = strcase.ToCamel(prefix)
				prefix = openapigen.UpperCaseFirstLetter(prefix)
				code, err := openapigen.ParseOperation(prefix, operation)
				if err != nil {
					if code != "" {
						fmt.Println("=========================================================")
						fmt.Println(code)
						fmt.Println("=========================================================")
					}

					fmt.Printf("ERROR: failed to parse operation %q: %v\n", prefix, err)
					os.Exit(1)
				}

				fileName := fmt.Sprintf("%s.go", prefix)
				filePath := path.Join(outDir, fileName)
				err = saveFile(filePath, code)
				checkError(err, "failed to save file %q", filePath)
			}
		}
	}
}

func saveUtils(outDir string) {
	src := `
		package server

		import (
			"encoding/json"
			"io"

			"github.com/gin-gonic/gin"
		)

		func parseBody(c *gin.Context, target any) error {
			bs, err := io.ReadAll(c.Request.Body)
			if err != nil {
				return nil
			}

			err = json.Unmarshal(bs, target)
			if err != nil {
				return err
			}

			return nil
		}
	`

	bs, err := format.Source([]byte(src))
	checkError(err, "failed to format utils source code")

	src = string(bs)
	filePath := path.Join(outDir, "apigenutils.go")
	err = saveFile(filePath, src)
	checkError(err, "failed to save file %q", filePath)
}

func checkError(err error, format string, a ...any) {
	if err != nil {
		msg := fmt.Sprintf(format, a...)
		fmt.Printf("ERROR: %s: %v\n", msg, err)
		os.Exit(1)
	}
}

func saveFile(path string, data string) error {
	err := os.WriteFile(path, []byte(data), os.ModePerm)
	if err != nil {
		return err
	}

	fmt.Printf("saved to: %s\n", path)
	return nil
}
