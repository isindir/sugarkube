package convert

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapInterfaceInterfaceToMapStringString(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		input         map[interface{}]interface{}
		expectValues  map[string]string
		expectedError bool
	}{
		{
			name: "good_conversion_string_keys",
			desc: "check converting expected input works",
			input: map[interface{}]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
			},
			expectValues: map[string]string{
				"testStr":   "hello",
				"testInt":   "3",
				"testFloat": "1.11",
				"testBool":  "true",
			},
			expectedError: false,
		},
		{
			name: "good_conversion_non_string_keys",
			desc: "check converting expected input works",
			input: map[interface{}]interface{}{
				3:   "hello",
				1.2: "world",
			},
			expectValues: map[string]string{
				"3":   "hello",
				"1.2": "world",
			},
			expectedError: false,
		},
		{
			name: "error_converting_sub_map",
			desc: "check converting map with sub-map causes an error",
			input: map[interface{}]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
				"sub": map[interface{}]interface{}{
					"subStr": "world",
				},
			},
			expectValues:  nil,
			expectedError: true,
		},
		{
			name: "error_converting_sub_array",
			desc: "check converting map with sub-array causes an error",
			input: map[interface{}]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
				"sub": []string{
					"subStr1",
					"subStr2",
					"subStr3",
				},
			},
			expectValues:  nil,
			expectedError: true,
		},
	}

	for _, test := range tests {
		result, err := MapInterfaceInterfaceToMapStringString(test.input)
		if test.expectedError {
			assert.NotNil(t, err)
			assert.Nil(t, result)
		} else {
			assert.Equal(t, test.expectValues, result, "unexpected conversion result for %s", test.name)
			assert.Nil(t, err)
		}
	}
}

func TestMapInterfaceInterfaceToMapStringInterface(t *testing.T) {
	tests := []struct {
		name          string
		desc          string
		input         map[interface{}]interface{}
		expectValues  map[string]interface{}
		expectedError bool
	}{
		{
			name: "good_conversion_string_keys",
			desc: "check converting expected input works",
			input: map[interface{}]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
			},
			expectValues: map[string]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
			},
			expectedError: false,
		},
		{
			name: "good_conversion_non_string_keys",
			desc: "check converting expected input works",
			input: map[interface{}]interface{}{
				3:   "hello",
				1.2: "world",
			},
			expectValues: map[string]interface{}{
				"3":   "hello",
				"1.2": "world",
			},
			expectedError: false,
		},
		{
			name: "good_convertion_sub_map",
			desc: "check converting map with sub-map works",
			input: map[interface{}]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
				"sub": map[interface{}]interface{}{
					"subStr": "world",
				},
			},
			expectValues: map[string]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
				"sub": map[interface{}]interface{}{
					"subStr": "world",
				},
			},
			expectedError: false,
		},
		{
			name: "good_converting_sub_array",
			desc: "check converting map with sub-array works",
			input: map[interface{}]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
				"sub": []string{
					"subStr1",
					"subStr2",
					"subStr3",
				},
			},
			expectValues: map[string]interface{}{
				"testStr":   "hello",
				"testInt":   3,
				"testFloat": 1.11,
				"testBool":  true,
				"sub": []string{
					"subStr1",
					"subStr2",
					"subStr3",
				},
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		result, err := MapInterfaceInterfaceToMapStringInterface(test.input)
		if test.expectedError {
			assert.NotNil(t, err)
			assert.Nil(t, result)
		} else {
			assert.Equal(t, test.expectValues, result, "unexpected conversion result for %s", test.name)
			assert.Nil(t, err)
		}
	}
}
