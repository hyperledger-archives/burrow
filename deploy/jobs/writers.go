package jobs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// [zr] this should go (currently used by the nameReg writer)
// WriteJobResultCSV takes two strings and writes those to the delineated log
// file, which is currently deploy.log in the same directory as the deploy.yaml
func WriteJobResultCSV(name, result string) error {

	pwd, _ := os.Getwd()
	logFile := filepath.Join(pwd, "jobs_output.csv")

	var file *os.File
	var err error

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		file, err = os.Create(logFile)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	defer file.Close()

	text := fmt.Sprintf("%s,%s\n", name, result)
	_, err = file.WriteString(text)

	return err
}

func WriteJobResultJSON(results map[string]interface{}, logFile string) error {

	file, err := os.Create(logFile)
	if err != nil {
		return err
	}
	defer file.Close()

	res, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	if _, err = file.Write(res); err != nil {
		return err
	}

	return nil
}
