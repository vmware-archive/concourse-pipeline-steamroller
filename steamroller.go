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

// interpreter represents the information required to invoke an interpreter.
type interpreter struct {
	// The path to the interpreter.
	Path string
	// The args used when invoking the interpreter with a script as an arg.
	Args []string
}

// interpreters maps file extensions to interpreters.
var interpreters = map[string]interpreter{
	"":    {"sh", []string{"-c"}},
	".sh": {"sh", []string{"-c"}},
	".rb": {"ruby", []string{"-e"}},
	".py": {"python", []string{"-c"}},
	".js": {"node", []string{"-e"}},
}

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

				var taskConfig atc.TaskConfig
				err = yaml.Unmarshal(taskConfigBytes, &taskConfig)
				if err != nil {
					log.Fatalf("failed to unmarshal task config at %s: %s", step.TaskConfigPath, err)
				}

				step.TaskConfig = &taskConfig
				step.TaskConfigPath = ""

				path := step.TaskConfig.Run.Path
				scriptBytes, err := loadBytes(files, path)
				if err != nil {
					log.Fatalf("failed to read task config at %s: %s", path, err)
				}

				interpreter := interpreters[filepath.Ext(path)]
				args := append([]string{}, interpreter.Args...)
				step.TaskConfig.Run.Args = append(args, string(scriptBytes))
				step.TaskConfig.Run.Path = interpreter.Path
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
