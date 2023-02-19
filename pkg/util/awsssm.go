package util

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/rs/zerolog/log"
	"os"
	"reflect"
	"sort"
	"strings"
)

var (
	trueBool        = true
	parameterType   = "SecureString"
	ErrNameNotFound = errors.New("Name not found")
)

type Flags struct {
	Export               bool
	InJson               bool
	OutJson              bool
	Upper                bool
	Quote                bool
	Dry                  bool
	Recursive            bool
	PrefixPath           bool
	PrefixNormalizedPath bool
}

type AWSSSM struct {
	session *session.Session
	SSM     ssmiface.SSMAPI
}

func IsPath(param *string) (bool, error) {
	if strings.Contains(*param, ",") {
		log.Error().Msgf("Parameter is unsplitted, contains a colon: %s", *param)
		return false, errors.New("Parameter is unsplitted, contains a colon")
	}
	if strings.Contains(*param, "/") {
		return true, nil
	}
	return false, nil
}

func SplitParams(param *string) []string {
	if *param != "" {
		return strings.Split(*param, ",")
	}

	return []string{}
}

func NewSSM() *AWSSSM {
	// initialize aws SSM
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	SSM := ssm.New(session)

	return &AWSSSM{
		session: session,
		SSM:     SSM,
	}
}

func getNameAndValue(param *ssm.Parameter, flags Flags) (string, string, error) {
	if flags.PrefixPath {
		return getUpper(*param.Name, flags), *param.Value, nil
	} else if flags.PrefixNormalizedPath {
		prefixPath := getUpper(*param.Name, flags)
		// remove first occurrence
		normalizedPath := strings.Replace(prefixPath, "/", "", 1)
		normalizedPath = strings.ReplaceAll(normalizedPath, "/", "_")
		return normalizedPath, *param.Value, nil
	} else {
		split := strings.Split(*param.Name, "/")
		name := split[len(split)-1]
		return getUpper(name, flags), *param.Value, nil
	}
}

func chunkParamNames(paramNames []*string, chunkSize int) [][]*string {
	var chunks [][]*string
	for i := 0; i < len(paramNames); i += chunkSize {
		end := i + chunkSize
		if end > len(paramNames) {
			end = len(paramNames)
		}

		chunks = append(chunks, paramNames[i:end])
	}

	return chunks
}

func (f *AWSSSM) GetParametersByPath(paths []string, flags Flags) (map[string]string, error) {
	params := make(map[string]string)

	// retrieve params for all paths
	for _, path := range paths {
		log.Debug().Msgf("Retrieving Path: %s", path)

		done := false
		var nextToken string
		for !done {
			input := &ssm.GetParametersByPathInput{
				Path:           &path,
				Recursive:      &flags.Recursive,
				WithDecryption: &trueBool,
			}

			if nextToken != "" {
				input.SetNextToken(nextToken)
			}

			output, err := f.SSM.GetParametersByPath(input)
			if err != nil {
				return params, err
			}
			log.Debug().Msgf("Retrieved Parameters for path %s: %s", path, output.Parameters)
			if len(output.Parameters) == 0 {
				// if no parameters are found, try to get the parameter as a single value
				inputSingle := &ssm.GetParameterInput{
					Name:           &path,
					WithDecryption: &trueBool,
				}
				outputSingle, err := f.SSM.GetParameter(inputSingle)
				if err != nil {
					// if this also fails, no path or parameter on path exists
					log.Error().Msgf("No names found for path: %s", path)
					return params, ErrNameNotFound
				}
				nameSingle, value, _ := getNameAndValue(outputSingle.Parameter, flags)
				log.Debug().Msgf("Retrieved Parameter for %s: %s", path, nameSingle)
				params[nameSingle] = value
				break
			}

			for _, param := range output.Parameters {
				name, value, _ := getNameAndValue(param, flags)
				log.Debug().Msgf("Name: %s Value %s", name, value)
				params[name] = value
			}

			// if nextToken has a value, there are more parameters to fetch. maximum is 10 parameters at a time.
			if output.NextToken != nil {
				nextToken = *output.NextToken
			} else {
				done = true
			}
		}
	}

	return params, nil
}

func (f *AWSSSM) GetParameters(ssmnames []*string, flags Flags) (map[string]string, error) {
	params := make(map[string]string)

	// GetParameters only supports at max of 10 params
	chunks := chunkParamNames(ssmnames, 10)

	// retrieve listed param names
	for _, chunk := range chunks {
		chunkNames := ""
		for _, name := range chunk {
			log.Debug().Msgf("Retrieving Name: %s", *name)
			chunkNames += fmt.Sprintf("%s ", *name)
		}

		input := &ssm.GetParametersInput{
			Names:          chunk,
			WithDecryption: &trueBool,
		}

		output, err := f.SSM.GetParameters(input)
		if err != nil {
			return params, err
		}
		if len(output.Parameters) == 0 {
			log.Error().Msgf("None of the Names was found: %s", chunkNames)
			return params, ErrNameNotFound
		}
		log.Debug().Msgf("Retrieved Parameters: %s", output.Parameters)

		for _, param := range output.Parameters {
			name, value, _ := getNameAndValue(param, flags)
			log.Debug().Msgf("NAME: %s VALUE: %s", name, value)
			params[name] = value
		}
	}

	return params, nil
}

func (f *AWSSSM) ReadParametersFromFile(fileName string, path string, flags Flags) (map[string]string, error) {
	params := make(map[string]string)

	if path != "" {
		isPath, err := IsPath(&path)
		if err != nil {
			log.Error().Msg(err.Error())
			return params, err
		}
		if !isPath {
			log.Error().Msgf("Target is not a path: %s", path)
			return params, errors.New("Target is not a path")
		}
	}

	if flags.InJson {
		dat, err := os.ReadFile(fileName)
		if err != nil {
			log.Error().Msg(err.Error())
			return params, err
		}
		params, err = ExpandJson(string(dat))
		if err != nil {
			log.Error().Msg(err.Error())
			return params, err
		}
	} else {
		file, err := os.Open(fileName)
		if err != nil {
			log.Error().Msg(err.Error())
			return params, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			log.Debug().Msgf("READ LINE: %s", scanner.Text())
			if strings.Index(scanner.Text(), "=") < 1 {
				log.Info().Msgf("Ignoring line: %s", scanner.Text())
			} else {
				data := strings.SplitN(scanner.Text(), "=", 2)
				name := data[0]
				value := data[1]
				if name != "" {
					log.Debug().Msgf("NAME: %s VALUE: %s", name, value)
					params[name] = value
				} else {
					log.Info().Msgf("Ignoring line: %s", scanner.Text())
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Error().Msg(err.Error())
			return params, err
		}
	}
	return params, nil
}

func (f *AWSSSM) SaveParametersFromFile(fileName string, basePath string, flags Flags) error {
	params, err := f.ReadParametersFromFile(fileName, basePath, flags)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	if flags.Dry {
		prefix := ""
		if basePath != "" {
			prefix = basePath + "/"
		}
		result := OutputParamsAsString(params, prefix, flags)
		fmt.Println("### Dry run, not saving, this would have been set:")
		fmt.Println(result)
		return nil
	}
	return f.SaveParameters(params, basePath)
}

func (f *AWSSSM) SaveParameters(params map[string]string, basePath string) error {
	var names []string
	for param := range params {
		names = append(names, param)
	}
	sort.Strings(names)
	for _, rawName := range names {
		paramName := rawName
		// construct a path if neccessary
		if basePath != "" {
			paramName = fmt.Sprintf("%s/%s", basePath, rawName)
		}
		value := params[rawName]
		fmt.Printf("%s=%s\n", paramName, value)
		input := &ssm.PutParameterInput{
			Name:      &paramName,
			Value:     &value,
			Overwrite: &trueBool,
			Type:      &parameterType,
		}

		output, err := f.SSM.PutParameter(input)
		if err != nil {
			return err
		}
		log.Info().Msgf("Output: %s", output)
	}

	return nil
}

func GetSortedNamesFromParams(params map[string]string) []string {
	var names []string
	for param := range params {
		names = append(names, param)
	}
	sort.Strings(names)
	return names
}

// outputs the parameters as string sorted by name
func OutputParamsAsString(params map[string]string, prefix string, flags Flags) string {
	names := GetSortedNamesFromParams(params)
	var result = ""
	for _, name := range names {
		if flags.Quote {
			result += fmt.Sprintf("%s%s=\"%s\"\n", prefix, name, params[name])
		} else {
			result += fmt.Sprintf("%s%s=%s\n", prefix, name, params[name])
		}
	}
	return result
}

func ExpandJson(value string) (map[string]string, error) {
	result := make(map[string]string)

	log.Debug().Str("json", value).Msg("ExpandJson")
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(value), &jsonMap)
	if err != nil {
		log.Error().Msgf("Error unmarshalling json: %s", err.Error())
		return nil, err
	}
	for jkey := range jsonMap {
		valueType := reflect.TypeOf(jsonMap[jkey])
		log.Debug().Str("jkey", jkey).Interface("jsonMap[jkey]", jsonMap[jkey]).Interface("valueType", valueType).Msg("ExpandJson jsonMap")
		switch jsonMap[jkey].(type) {
		case int, int8, int16, int32, int64:
			result[jkey] = fmt.Sprintf("%v", jsonMap[jkey])
		case float32, float64:
			// Integer "0" and float "0.0" result in a float type and hence are
			// indistinguishable here, and output as integer
			result[jkey] = fmt.Sprintf("%v", jsonMap[jkey])
		case bool:
			result[jkey] = fmt.Sprintf("%t", jsonMap[jkey])
		default:
			result[jkey] = fmt.Sprintf("%s", jsonMap[jkey])
		}
	}
	return result, nil
}

func ExpandJsonParams(params map[string]string, flags Flags) (map[string]string, error) {
	result := make(map[string]string)

	for name, value := range params {
		log.Debug().Str("name", name).Msg("ExpandJsonParams")

		valueMap, err := ExpandJson(value)
		if err != nil {
			log.Error().Msgf("Error unmarshalling ssm parameter: %s / %s", name, err.Error())
			return nil, err
		}
		for jkey := range valueMap {
			resultKey := getUpper(jkey, flags)
			log.Debug().Msgf("valueMap: %s = %s", jkey, valueMap[jkey])
			result[resultKey] = fmt.Sprintf("%s", valueMap[jkey])
		}
	}
	return result, nil
}

func getUpper(param string, flags Flags) string {
	if flags.Upper {
		return strings.ToUpper(param)
	}
	return param
}

func (f *AWSSSM) GetParams(paramstring *string, flags Flags) (map[string]string, error) {
	results := make(map[string]string)

	params := SplitParams(paramstring)
	paramNames := make([]*string, 0)
	pathNames := make([]string, 0)

	for index := range params {
		parameter := params[index]
		isPath, err := IsPath(&parameter)
		if err != nil {
			log.Error().Msg(err.Error())
			return results, err
		}
		if isPath {
			pathNames = append(pathNames, parameter)
			log.Debug().Msgf("Parameter Path: %s", parameter)
		} else {
			paramNames = append(paramNames, &parameter)
			log.Debug().Msgf("Parameter Name: %s", parameter)
		}
	}

	pathResults, err := f.GetParametersByPath(pathNames, flags)
	if err != nil {
		log.Error().Msg(err.Error())
		return results, err
	}

	for name, value := range pathResults {
		log.Debug().Msgf("Name: %s Value %s", name, value)
		results[name] = value
	}

	singleResults, err := f.GetParameters(paramNames, flags)
	if err != nil {
		log.Error().Msg(err.Error())
		return results, err
	}

	for name, value := range singleResults {
		log.Debug().Msgf("Name: %s Value %s", name, value)
		results[name] = value
	}

	if flags.InJson {
		results, err = ExpandJsonParams(results, flags)
		if err != nil {
			log.Error().Msg(err.Error())
			return results, err
		}
	}
	return results, nil
}

func (f *AWSSSM) GetOutputString(results map[string]string, flags Flags) (string, error) {
	if flags.Export && flags.OutJson {
		log.Error().Msg("--export and --outjson can not be used together")
		return "", errors.New("export and outjson can not be used together")
	}

	var result string
	if flags.OutJson {
		json, _ := json.MarshalIndent(results, "", "  ")
		result = fmt.Sprint(string(json))
	} else {
		if flags.Export {
			result = OutputParamsAsString(results, "export ", flags)
		} else {
			result = OutputParamsAsString(results, "", flags)
		}
	}
	return result, nil
}
