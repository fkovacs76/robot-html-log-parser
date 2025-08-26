package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func listTests(elemList interface{}, strList []string) {
	testArr, ok := elemList.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}
	fmt.Println("Nr. of tests:", len(testArr))

	for i, val := range testArr {
		fmt.Printf("Index: %d, Value: %#v\n", i, val)
		testObj, ok := val.([]interface{})
		if !ok {
			fmt.Println("Unexpected type")
			return
		}

		fmt.Println("Name: ", strList[int(testObj[0].(float64))])
		fmt.Println("Doc: ", strList[int(testObj[2].(float64))])

		listKeyWords(testObj[5], strList)
	}
}

func listKeyWords(elemList interface{}, strList []string) {
	keyWordArr, ok := elemList.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}
	fmt.Println("Nr. of keywords:", len(keyWordArr))

	for i, val := range keyWordArr {
		fmt.Printf("Index: %d, Value: %#v\n", i, val)
		keyWordObj, ok := val.([]interface{})
		if !ok {
			fmt.Println("Unexpected type")
			return
		}

		fmt.Println("Name: ", strList[int(keyWordObj[1].(float64))])
		fmt.Println("Libname: ", strList[int(keyWordObj[2].(float64))])
		fmt.Println("Args: ", strList[int(keyWordObj[5].(float64))])
	}
}

// extractSuiteData extracts the JSON data from window.output["suite"] assignment
func extractSuiteData(htmlContent string) (string, error) {
	re := regexp.MustCompile(`window\.output\["suite"\]\s*=\s*(\[.*?\]);`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find window.output[\"suite\"] assignment")
	}
	return matches[1], nil
}

// extractStringsData extracts the JSON data from window.output["strings"] concat assignment
func extractStringsData(htmlContent string) (string, error) {
	re := regexp.MustCompile(`window\.output\["strings"\]\.concat\((\[.*?\])\);`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find window.output[\"strings\"] concat assignment")
	}
	return matches[1], nil
}

// readHTMLFile reads the HTML file and extracts the required JSON data
func readHTMLFile(filename string) (string, string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("error reading file: %v", err)
	}

	htmlContent := string(content)

	suiteData, err := extractSuiteData(htmlContent)
	if err != nil {
		return "", "", err
	}

	stringsData, err := extractStringsData(htmlContent)
	if err != nil {
		return "", "", err
	}

	return suiteData, stringsData, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run extract.go <html_file>")
		os.Exit(1)
	}

	htmlFile := os.Args[1]

	data, output_strings, err := readHTMLFile(htmlFile)
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		os.Exit(1)
	}

	cleaned := strings.ReplaceAll(output_strings, `\x3c`, "<")

	var result interface{}
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var outputArr []string
	err = json.Unmarshal([]byte(cleaned), &outputArr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("%#v\n", outputArr)

	fmt.Printf("%#v\n", result)

	arr, ok := result.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}

	fmt.Printf("%#v\n", arr[5])
	fmt.Println("Suites")
	fmt.Printf("%#v\n", arr[6])
	fmt.Println("Tests")
	fmt.Printf("%#v\n", arr[7])
	fmt.Println("Keywords")
	fmt.Printf("%#v\n", arr[8])
	fmt.Println("Don't know")
	fmt.Printf("%#v\n", arr[9])

	listTests(arr[7], outputArr)

}
