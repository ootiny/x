package x

import (
	"embed"
	"encoding/json"
	"maps"
	"os"
	"strings"
)

func IsFileExists(efs *embed.FS, path string) bool {
	_, err := efs.ReadFile(path)
	return err == nil
}

func fnListJsonTasks(efs *embed.FS, dir string, subDir string) ([]string, error) {
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

func ListJsonTasks(efs *embed.FS, dir string) ([]string, error) {
	return fnListJsonTasks(efs, dir, "")
}

func LoadTaskConfig(
	config any,
	efs *embed.FS,
	taskName string,
	overrideConfig map[string]any,
) error {
	var taskJSON []byte
	var err error
	if strings.HasSuffix(taskName, ".json") {
		taskJSON, err = os.ReadFile(taskName)
	} else if efs == nil {
		return Errorf("task %s: not found", taskName)
	} else if !IsFileExists(efs, taskName+".json") {
		return Errorf(`task "%s" not found`, taskName)
	} else {
		taskJSON, err = efs.ReadFile(taskName + ".json")
	}

	if err != nil {
		return err
	}

	baseConfig := map[string]any{}
	if err := json.Unmarshal(taskJSON, &baseConfig); err != nil {
		return Errorf(`task %s: failed to unmarshal base config: %v`, taskName, err)
	}

	maps.Copy(baseConfig, overrideConfig)

	if mergedConfig, err := json.Marshal(baseConfig); err != nil {
		return Errorf(`task %s: failed to marshal merged config: %v`, taskName, err)
	} else if err := json.Unmarshal(mergedConfig, config); err != nil {
		return Errorf(`task %s: failed to unmarshal merged config: %v`, taskName, err)
	} else {
		return nil
	}
}
