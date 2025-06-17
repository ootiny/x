package x

import (
	"embed"
	"strings"
)

func fnListJsonTasks(efs embed.FS, dir string, subDir string) ([]string, error) {
	var result []string

	if entries, err := efs.ReadDir(Ternary(subDir == "", dir, dir+"/"+subDir)); err != nil {
		return nil, err
	} else {
		for _, entry := range entries {
			subPath := Ternary(subDir == "", entry.Name(), subDir+"/"+entry.Name())
			if entry.IsDir() {
				if subFiles, err := fnListJsonTasks(efs, dir, subPath); err != nil {
					return nil, err
				} else {
					result = append(result, subFiles...)
				}
			} else if strings.HasSuffix(entry.Name(), ".json") {
				result = append(result, strings.TrimSuffix(subPath, ".json"))
			}
		}

		return result, nil
	}
}

func ListJsonTasks(efs embed.FS, dir string) ([]string, error) {
	return fnListJsonTasks(efs, dir, "")
}
