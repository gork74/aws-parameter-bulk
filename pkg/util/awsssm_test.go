package util

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/rs/zerolog/log"
	"testing"
)

type MockSSM struct {
	ssmiface.SSMAPI
	err error
}

func nameString(parameter ssm.GetParametersInput) string {
	result := "["
	for i, param := range parameter.Names {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s", *param)
	}
	return result + "]"
}

func (sp *MockSSM) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	output := new(ssm.GetParametersOutput)
	log.Info().Msgf("%s", nameString(*input))
	if nameString(*input) == "[One1]" {
		name1 := "One1"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("OneVal1")})
	}
	if nameString(*input) == "[One2]" {
		name1 := "One2"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("OneVal2")})
	}
	if nameString(*input) == "[One1, One2]" {
		name1 := "One1"
		name2 := "One2"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("OneVal1")})
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name2, Value: aws.String("OneVal2")})
	}
	if nameString(*input) == "[Three1, Three2]" {
		name1 := "Three1"
		name2 := "Three2"
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name1, Value: aws.String("ThreeVal1")})
		output.Parameters = append(output.Parameters, &ssm.Parameter{Name: &name2, Value: aws.String("ThreeVal2")})
	}
	return output, sp.err
}

func (sp *MockSSM) GetParametersByPath(input *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	output := new(ssm.GetParametersByPathOutput)
	params := make([]*ssm.Parameter, 0)
	if *input.Path == "/path" {
		params = append(params, &ssm.Parameter{Name: aws.String("One1"), Value: aws.String("OneVal1")})
	}
	if *input.Path == "/path2" {
		params = append(params, &ssm.Parameter{Name: aws.String("One1"), Value: aws.String("OneVal1")})
		params = append(params, &ssm.Parameter{Name: aws.String("One2"), Value: aws.String("OneVal2")})
	}
	output.Parameters = params
	return output, sp.err
}

func Test_IsPath(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "/path",
			want: true,
		},
		{
			name: "nopath",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsPath(&tt.name)
			t.Logf("%t", result)
			if err != nil {
				t.Error("Error in IsPath")
			}
			if result != tt.want {
				t.Errorf("Expected %t but got %t", tt.want, result)
			}
		})
	}
}

func Test_SplitParams(t *testing.T) {
	tests := []struct {
		name string
		want interface{}
	}{
		{
			name: "/path",
			want: [...]string{"/path"},
		},
		{
			name: "nopath",
			want: [...]string{"nopath"},
		},
		{
			name: "/path,single",
			want: [...]string{"/path", "single"},
		},
		{
			name: "/path,single,/path2",
			want: [...]string{"/path", "single", "/path2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitParams(&tt.name)
			t.Logf("%s", result)
			t.Logf("%s", tt.want)
			resultStr := fmt.Sprintf("%s", result)
			wantStr := fmt.Sprintf("%s", tt.want)
			if resultStr != wantStr {
				t.Errorf("Expected %s but got %s", tt.want, result)
			}
		})
	}
}

func Test_ExpandJson(t *testing.T) {
	value := `{"Name1":"Alice","Name2":"Bob"}`
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Name1",
			want: "Alice",
		},
		{
			name: "Name2",
			want: "Bob",
		},
	}
	result, err := ExpandJson(value)
	if err != nil {
		t.Error("Error expanding json")
	}
	t.Logf("%s", result)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result[tt.name] != tt.want {
				t.Errorf("Expected %s but got %s", tt.want, result[tt.name])
			}
		})
	}
}

func Test_ExpandJsonParams(t *testing.T) {
	input := make(map[string]string)
	input["one"] = `{"One1":"OneVal1","One2":"OneVal2"}`
	input["two"] = `{"Two1":"TwoVal1","Two2":"TwoVal2"}`
	tests := []struct {
		name string
		want string
	}{
		{
			name: "One1",
			want: "OneVal1",
		},
		{
			name: "One2",
			want: "OneVal2",
		},
		{
			name: "Two1",
			want: "TwoVal1",
		},
		{
			name: "Two2",
			want: "TwoVal2",
		},
	}
	result, err := ExpandJsonParams(input)
	if err != nil {
		t.Error("Error expanding json Params")
	}
	t.Logf("%s", result)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result[tt.name] != tt.want {
				t.Errorf("Expected %s but got %s", tt.want, result[tt.name])
			}
		})
	}
}

func Test_GetParams(t *testing.T) {
	input := make(map[string]string)
	input["One"] = `{"One1":"OneVal1","One2":"OneVal2"}`
	input["Two"] = `{"Two1":"TwoVal1","Two2":"TwoVal2"}`
	tests := []struct {
		params  string
		injson  bool
		outjson bool
		export  bool
		upper   bool
		quote   bool
		want    string
	}{
		{
			params:  "/path",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   true,
			quote:   false,
			want:    "ONE1=OneVal1\n",
		},
		{
			params:  "/path",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   false,
			quote:   false,
			want:    "One1=OneVal1\n",
		},
		{
			params:  "/path,/path2",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   true,
			quote:   false,
			want:    "ONE1=OneVal1\nONE2=OneVal2\n",
		},
		{
			params:  "One1",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   true,
			quote:   false,
			want:    "ONE1=OneVal1\n",
		},
		{
			params:  "One1,One2",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   true,
			quote:   false,
			want:    "ONE1=OneVal1\nONE2=OneVal2\n",
		},
		{
			params:  "/path,Three1,Three2,/path2",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   true,
			quote:   false,
			want:    "ONE1=OneVal1\nONE2=OneVal2\nTHREE1=ThreeVal1\nTHREE2=ThreeVal2\n",
		},
		{
			params:  "/path,Three1,Three2,/path2",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   false,
			quote:   false,
			want:    "One1=OneVal1\nOne2=OneVal2\nThree1=ThreeVal1\nThree2=ThreeVal2\n",
		},
		{
			params:  "/path,Three1,Three2,/path2",
			injson:  false,
			outjson: false,
			export:  false,
			upper:   false,
			quote:   true,
			want:    "One1=\"OneVal1\"\nOne2=\"OneVal2\"\nThree1=\"ThreeVal1\"\nThree2=\"ThreeVal2\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.params, func(t *testing.T) {
			ssmClient := NewSSM()
			ssmClient.SSM = &MockSSM{
				err: nil, //errors.New("my custom error"),
			}
			result, err := ssmClient.GetParams(&tt.params, tt.injson, tt.upper)
			if err != nil {
				t.Error("Error in GetParams")
			}
			output, err := ssmClient.GetOutputString(result, tt.outjson, tt.export, tt.quote)
			outputStr := fmt.Sprintf("%s", output)
			wantStr := fmt.Sprintf("%s", tt.want)
			if outputStr != wantStr {
				t.Errorf("Expected '%s' but got '%s'", tt.want, output)
			}
		})
	}
}
