package cfn_test

import (
	"reflect"
	"testing"

	"github.com/aws-cloudformation/rain/cfn"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

var testMap = map[string]interface{}{
	"Parameters": map[string]interface{}{
		"Name": map[string]interface{}{
			"Type": "String",
		},
	},
	"Resources": map[string]interface{}{
		"Bucket": map[string]interface{}{
			"Type": "AWS::S3::Bucket",
			"Properties": map[string]interface{}{
				"BucketName": map[string]interface{}{
					"Ref": "Name",
				},
			},
		},
	},
	"Outputs": map[string]interface{}{
		"BucketArn": map[string]interface{}{
			"Value": map[string]interface{}{
				"GetAtt": []interface{}{
					"Bucket",
					"Arn",
				},
			},
		},
	},
}

var testTemplate = cfn.NewTemplate(testMap)

func TestGraph(t *testing.T) {
	graph := testTemplate.Graph()

	actual := graph.Nodes()
	expected := []interface{}{
		cfn.Element{"BucketArn", "Outputs"},
		cfn.Element{"Name", "Parameters"},
		cfn.Element{"Bucket", "Resources"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Template graph is wrong:\n%#v\n!=\n%#v\n", expected, actual)
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		path     []interface{}
		expected interface{}
	}{
		{[]interface{}{}, testMap},
		{[]interface{}{"Parameters"}, testMap["Parameters"]},
		{[]interface{}{"Parameters", "Name"}, testMap["Parameters"].(map[string]interface{})["Name"]},
		{[]interface{}{"Outputs", "BucketArn", "Value", "GetAtt", 0}, "Bucket"},
		{[]interface{}{"Outputs", "BucketArn", "Value", "GetAtt", 1}, "Arn"},
	}

	for _, testCase := range testCases {
		actual, err := testTemplate.Get(testCase.path)
		if err != nil {
			t.Error(err)
		}

		actualYaml, err := yaml.Marshal(actual)
		if err != nil {
			t.Error(err)
		}

		expectedYaml, err := yaml.Marshal(testCase.expected)
		if err != nil {
			t.Error(err)
		}

		if d := cmp.Diff(actualYaml, expectedYaml); d != "" {
			t.Error(d)
		}
	}
}

func TestSet(t *testing.T) {
	testCases := []struct {
		path  []interface{}
		value interface{}
	}{
		{[]interface{}{"Foo"}, []interface{}{}},
		{[]interface{}{"Foo", 0}, "Bar"},
		{[]interface{}{"Foo", 1}, map[string]interface{}{}},
		{[]interface{}{"Foo", 1, "Baz"}, "Quux"},
	}

	expectedYaml, _ := yaml.Marshal(map[string]interface{}{
		"Foo": []interface{}{
			"Bar",
			map[string]interface{}{
				"Baz": "Quux",
			},
		},
	})

	actual := cfn.NewTemplate(map[string]interface{}{})

	for _, testCase := range testCases {
		err := actual.Set(testCase.path, testCase.value)
		if err != nil {
			t.Error(err)
		}
	}

	actualYaml, err := yaml.Marshal(actual.Map())
	if err != nil {
		t.Error(err)
	}

	if d := cmp.Diff(actualYaml, expectedYaml); d != "" {
		t.Error(d)
	}
}

func TestSetCreate(t *testing.T) {
	testCases := []struct {
		path  []interface{}
		value interface{}
	}{
		{[]interface{}{"Foo", 0}, "Bar"},
		{[]interface{}{"Foo", 1, "Baz"}, "Quux"},
	}

	expectedYaml, _ := yaml.Marshal(map[string]interface{}{
		"Foo": []interface{}{
			"Bar",
			map[string]interface{}{
				"Baz": "Quux",
			},
		},
	})

	actual := cfn.NewTemplate(map[string]interface{}{})

	for _, testCase := range testCases {
		err := actual.Set(testCase.path, testCase.value)
		if err != nil {
			t.Error(err)
		}
	}

	actualYaml, err := yaml.Marshal(actual.Map())
	if err != nil {
		t.Error(err)
	}

	if d := cmp.Diff(actualYaml, expectedYaml); d != "" {
		t.Error(d)
	}
}
