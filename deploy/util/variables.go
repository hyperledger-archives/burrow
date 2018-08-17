package util

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"encoding/json"

	"unicode"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/def/rule"
	"github.com/hyperledger/burrow/execution/evm/abi"
	log "github.com/sirupsen/logrus"
)

func Variables(value interface{}) []*abi.Variable {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()
	var variables []*abi.Variable
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.Kind() == reflect.String {
			variables = append(variables, &abi.Variable{Name: lowerFirstCharacter(rt.Field(i).Name), Value: field.String()})
		}

	}
	return variables
}

func lowerFirstCharacter(name string) string {
	if name == "" {
		return name
	}
	bs := []byte(name)
	bs[0] = byte(unicode.ToLower(rune(bs[0])))
	return string(bs)
}

func PreProcessFields(value interface{}, do *def.Packages) (err error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.Kind() == reflect.String {
			str, err := PreProcess(field.String(), do)
			if err != nil {
				return err
			}
			field.SetString(str)
		}
	}
	return nil
}

func PreProcess(toProcess string, do *def.Packages) (string, error) {
	// Run through the replacement process for any placeholder matches
	for _, pm := range rule.MatchPlaceholders(toProcess) {
		log.WithField("match", toProcess).Debug("Replacement Match Found")

		// first parse the reserved words.
		if strings.Contains(pm.JobName, "block") {
			block, err := replaceBlockVariable(pm.Match, do)
			if err != nil {
				log.WithField("err", err).Error("Error replacing block variable.")
				return "", err
			}
			/*log.WithFields(log.Fields{
				"var": toProcess,
				"res": block,
			}).Debug("Fixing Variables =>")*/
			toProcess = strings.Replace(toProcess, pm.Match, block, 1)
			continue
		}

		// second we loop through the jobNames to do a result replace
		for _, job := range do.Package.Jobs {
			if pm.JobName == job.Name {
				if pm.VariableName != "" {
					for _, variable := range job.Variables {
						if variable.Name == pm.VariableName { //find the value we want from the bunch
							toProcess = strings.Replace(toProcess, pm.Match, variable.Value, 1)
							log.WithFields(log.Fields{
								"job":     pm.JobName,
								"varName": pm.VariableName,
								"result":  variable.Value,
							}).Debug("Fixing Inner Vars =>")
						}
					}
				} else {
					// If result is returned as string assume that rendering otherwise marshal to JSON
					result, ok := job.Result.(string)
					if !ok {
						bs, err := json.Marshal(job.Result)
						if err != nil {
							return "", fmt.Errorf("error marhsalling tx result in post processing: %v", err)
						}
						result = string(bs)
					}
					log.WithFields(log.Fields{
						"var": string(pm.JobName),
						"res": result,
					}).Debug("Fixing Variables =>")
					toProcess = strings.Replace(toProcess, pm.Match, result, 1)
				}
			}
		}
	}
	return toProcess, nil
}

func replaceBlockVariable(toReplace string, do *def.Packages) (string, error) {
	log.WithFields(log.Fields{
		"var": toReplace,
	}).Debug("Correcting $block variable")
	blockHeight, err := GetBlockHeight(do)
	block := itoaU64(blockHeight)
	log.WithField("=>", block).Debug("Current height is")
	if err != nil {
		return "", err
	}

	if toReplace == "$block" {
		log.WithField("=>", block).Debug("Replacement (=)")
		return block, nil
	}

	catchEr := regexp.MustCompile(`\$block\+(\d*)`)
	if catchEr.MatchString(toReplace) {
		height := catchEr.FindStringSubmatch(toReplace)[1]
		h1, err := strconv.Atoi(height)
		if err != nil {
			return "", err
		}
		h2, err := strconv.Atoi(block)
		if err != nil {
			return "", err
		}
		height = strconv.Itoa(h1 + h2)
		log.WithField("=>", height).Debug("Replacement (+)")
		return height, nil
	}

	catchEr = regexp.MustCompile(`\$block\-(\d*)`)
	if catchEr.MatchString(toReplace) {
		height := catchEr.FindStringSubmatch(toReplace)[1]
		h1, err := strconv.Atoi(height)
		if err != nil {
			return "", err
		}
		h2, err := strconv.Atoi(block)
		if err != nil {
			return "", err
		}
		height = strconv.Itoa(h1 - h2)
		log.WithField("=>", height).Debug("Replacement (-)")
		return height, nil
	}

	log.WithField("=>", toReplace).Debug("Replacement (unknown)")
	return toReplace, nil
}

func PreProcessInputData(function string, data interface{}, do *def.Packages, constructor bool) (string, []string, error) {
	var callDataArray []string
	var callArray []string
	if function == "" && !constructor {
		if reflect.TypeOf(data).Kind() == reflect.Slice {
			return "", []string{""}, fmt.Errorf("Incorrect formatting of deploy.yaml. Please update it to include a function field.")
		}
		function = strings.Split(data.(string), " ")[0]
		callArray = strings.Split(data.(string), " ")[1:]
		for _, val := range callArray {
			output, _ := PreProcess(val, do)
			callDataArray = append(callDataArray, output)
		}
	} else if data != nil {
		if reflect.TypeOf(data).Kind() != reflect.Slice {
			if constructor {
				log.Warn("Deprecation Warning: Your deploy job is currently using a soon to be deprecated way of declaring constructor values. Please remember to update your run file to store them as a array rather than a string. See documentation for further details.")
				callArray = strings.Split(data.(string), " ")
				for _, val := range callArray {
					output, _ := PreProcess(val, do)
					callDataArray = append(callDataArray, output)
				}
				return function, callDataArray, nil
			} else {
				return "", make([]string, 0), fmt.Errorf("Incorrect formatting of deploy.yaml file. Please update it to include a function field.")
			}
		}
		val := reflect.ValueOf(data)
		for i := 0; i < val.Len(); i++ {
			s := val.Index(i)
			var newString string
			switch s.Interface().(type) {
			case bool:
				newString = strconv.FormatBool(s.Interface().(bool))
			case int, int32, int64:
				newString = strconv.FormatInt(int64(s.Interface().(int)), 10)
			case []interface{}:
				var args []string
				for _, index := range s.Interface().([]interface{}) {
					value := reflect.ValueOf(index)
					var stringified string
					switch value.Kind() {
					case reflect.Int:
						stringified = strconv.FormatInt(value.Int(), 10)
					case reflect.String:
						stringified = value.String()
					}
					index, _ = PreProcess(stringified, do)
					args = append(args, stringified)
				}
				newString = "[" + strings.Join(args, ",") + "]"
				log.Debug(newString)
			default:
				newString = s.Interface().(string)
			}
			newString, _ = PreProcess(newString, do)
			callDataArray = append(callDataArray, newString)
		}
	}
	return function, callDataArray, nil
}

func PreProcessLibs(libs string, do *def.Packages) (string, error) {
	libraries, _ := PreProcess(libs, do)
	if libraries != "" {
		pairs := strings.Split(libraries, ",")
		libraries = strings.Join(pairs, " ")
	}
	log.WithField("=>", libraries).Debug("Library String")
	return libraries, nil
}

func GetReturnValue(vars []*abi.Variable) string {
	var result []string

	if len(vars) > 1 {
		for _, value := range vars {
			log.WithField("=>", []byte(value.Value)).Debug("Value")
			result = append(result, value.Value)
		}
		return "(" + strings.Join(result, ", ") + ")"
	} else if len(vars) == 1 {
		log.Debug("Debugging: ", vars[0].Value)
		return vars[0].Value
	} else {
		return ""
	}
}
