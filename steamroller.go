package steamroller

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/concourse/atc"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ResourceMap map[string]string `yaml:"resource_map"`
}

func Steamroll(files map[string]string, atcConfig atc.Config) (*atc.Config, error) {
	newConfig := atcConfig

	for i := range atcConfig.Jobs {
		newPlan := atcConfig.Jobs[i].Plan
		flattenPlanConfig(files, newPlan)

		newConfig.Jobs[i].Plan = newPlan
	}

	return &newConfig, nil
}

func flattenPlanConfig(files map[string]string, jobPlan []atc.PlanConfig) {
	for i, step := range jobPlan {
		switch {
		case step.Aggregate != nil:
			flattenPlanConfig(files, []atc.PlanConfig(*step.Aggregate))

		case step.Do != nil:
			flattenPlanConfig(files, []atc.PlanConfig(*step.Do))

		case step.Task != "":
			if step.TaskConfigPath != "" {
				if files == nil {
					log.Fatalf("empty resource map; cannot find %s", step.TaskConfigPath)
				}

				taskConfigBytes, err := loadBytes(files, step.TaskConfigPath)
				if err != nil {
					log.Fatalf("failed to read task config at %s: %s", step.TaskConfigPath, err)
				}

				var taskConfig atc.LoadTaskConfig
				err = yaml.Unmarshal(taskConfigBytes, &taskConfig)
				if err != nil {
					log.Fatalf("failed to unmarshal task config at %s: %s", step.TaskConfigPath, err)
				}

				step.TaskConfig = &taskConfig
				step.TaskConfigPath = ""

				scriptBytes, err := loadBytes(files, step.TaskConfig.Run.Path)
				if err != nil {
					log.Fatalf("failed to read task config at %s: %s", step.TaskConfig.Run.Path, err)
				}

				step.TaskConfig.Run.Args = []string{"-c", string(scriptBytes)}
				step.TaskConfig.Run.Path = "sh"
			}

			jobPlan[i] = step
		}
	}
}

func loadBytes(resourceMap map[string]string, path string) ([]byte, error) {
	resourceRoot := strings.Split(path, string(os.PathSeparator))[0]

	resourcePath, ok := resourceMap[resourceRoot]
	if !ok || resourcePath == "" {
		return nil, fmt.Errorf("no resource map provided for %s", path)
	}

	actualPath := filepath.Join(resourcePath, strings.Replace(path, resourceRoot, "", -1))

	return ioutil.ReadFile(actualPath)
}
