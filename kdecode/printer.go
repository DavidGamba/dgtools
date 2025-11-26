package main

import (
	"fmt"
	"io"
	"slices"

	"go.yaml.in/yaml/v4"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

func resourcePrint(outputFormat string, obj runtime.Object, w io.Writer) error {
	var printer printers.ResourcePrinter
	switch outputFormat {
	case "yaml":
		printer = &printers.YAMLPrinter{}
	case "json":
		printer = &printers.JSONPrinter{}
	default:
		return fmt.Errorf("unknown output format: %s", outputFormat)
	}
	printer = &printers.OmitManagedFieldsPrinter{Delegate: printer}
	p := printers.NewTypeSetter(scheme.Scheme).ToPrinter(printer)
	err := p.PrintObj(obj, w)
	if err != nil {
		return err
	}
	fmt.Fprintln(w)
	return nil
}

// Takes the Data and StringData portions of a k8s secret and sorts the keys, then it tries to parse the values as a pem cert
// Optionally limit output to a single given key
func SecretPrint(k8sSecret *v1.Secret, pem bool, key string) {
	// sort output
	dataKeys := []string{}
	for k := range k8sSecret.Data {
		dataKeys = append(dataKeys, k)
	}
	slices.Sort(dataKeys)
	stringDataKeys := []string{}
	for k := range k8sSecret.StringData {
		stringDataKeys = append(stringDataKeys, k)
	}
	slices.Sort(stringDataKeys)

	for _, k := range dataKeys {
		v := k8sSecret.Data[k]
		if (key != "" && k == key) || key == "" {
			fmt.Printf("%s=", k)
			if pem {
				info, err := ParseCert(string(v))
				if err != nil {
					fmt.Printf("%s\n", string(v))
					Logger.Printf("%s: %s\n", k, err)
				} else {
					fmt.Printf("\n%s\n", info)
				}
			} else {
				fmt.Printf("%s\n", string(v))
			}
		}
	}

	for _, k := range stringDataKeys {
		v := k8sSecret.StringData[k]
		if (key != "" && k == key) || key == "" {
			fmt.Printf("%s=", k)
			if pem {
				info, err := ParseCert(v)
				if err != nil {
					fmt.Printf("%s\n", v)
					Logger.Printf("%s\n", err)
				} else {
					fmt.Printf("\n%s\n", info)
				}
			} else {
				fmt.Printf("%s\n", v)
			}
		}
	}
}

// trimData recursively removes nil and empty fields
func trimData(value any) any {
	switch v := value.(type) {
	case map[string]any:
		// Recursively trim nested maps
		trimmed := make(map[string]any)
		for key, value := range v {
			trimmedValue := trimData(value)
			if trimmedValue != nil {
				trimmed[key] = trimmedValue
			}
		}
		// Return nil if map is empty after trimming
		if len(trimmed) == 0 {
			return nil
		}
		return trimmed
	case []any:
		// Recursively trim array elements
		trimmed := []any{}
		for _, item := range v {
			trimmedItem := trimData(item)
			if trimmedItem != nil {
				trimmed = append(trimmed, trimmedItem)
			}
		}
		// Return nil if array is empty after trimming
		if len(trimmed) == 0 {
			return nil
		}
		return trimmed
	case string:
		// Return nil for empty strings
		if v == "" {
			return nil
		}
		return v
	case nil:
		// Skip nil values
		return nil
	default:
		// Include all other values (numbers, booleans, etc.)
		return v
	}
}

// marshalTrim marshals an object to YAML removing nil and empty fields
func marshalTrim(obj any) ([]byte, error) {
	// First marshal to get the data, without this step we would have to reflect to iterate over the struct itself.
	// This first step converts the struct into a string representation of key values.
	data, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	// Unmarshal back into a Go data structure, except this time we only have to worry about maps of strings, arrays or leaf types.
	var simpleData any
	err = yaml.Unmarshal(data, &simpleData)
	if err != nil {
		return nil, err
	}

	// Trim empty fields
	trimmed := trimData(simpleData)

	// Marshal back to YAML
	return yaml.Marshal(trimmed)
}
