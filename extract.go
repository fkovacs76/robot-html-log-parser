package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const INDENT = 4

func returnIndent(level int) string {
	return strings.Repeat(" ", level*INDENT)
}

func listTests(elemList interface{}, strList []string) {
	testArr, ok := elemList.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}
	fmt.Println("Nr. of tests:", len(testArr))

	for _, val := range testArr {
		//fmt.Printf("Tests Index: %d, Value: %#v\n", i, val)
		testObj, ok := val.([]interface{})
		if !ok {
			fmt.Println("Unexpected type")
			return
		}

		fmt.Println("Name: ", strList[int(testObj[0].(float64))])
		fmt.Println("Doc: ", strList[int(testObj[2].(float64))])

		listKeyWords(testObj[5], strList, 0)
	}
}

// func listMessage(elemList interface{}, strList []string) {
// 	msgArr, ok := elemList.([]interface{})
// 	if !ok {
// 		fmt.Println("Unexpected type")
// 		return
// 	}
// 	fmt.Println("Nr. of messages:", len(msgArr))

// 	for i, val := range msgArr {
// 		fmt.Printf("Messages Index: %d, Value: %#v\n", i, val)
// 		msgObj, ok := val.([]interface{})
// 		if !ok {
// 			fmt.Println("Unexpected type")
// 			return
// 		}

// 		fmt.Println("Message: ", strList[int(msgObj[2].(float64))])
// 		//timestamp is index 0
// 		//index 1 is log levelin LEVELS array
// 	}
// }

func listKeyWords(elemList interface{}, strList []string, index int) {
	keyWordArr, ok := elemList.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}
	fmt.Println("Nr. of keywords:", len(keyWordArr))

	for _, val := range keyWordArr {
		//fmt.Printf("Keywords Index: %d, Value: %#v\n", i, val)
		keyWordObj, ok := val.([]interface{})
		if !ok {
			fmt.Println("Unexpected type")
			return
		}

		if len(keyWordObj) < 5 {
			//this is a message, not a keyword
			fmt.Println(returnIndent(index), "Message: ", strList[int(keyWordObj[2].(float64))])
			fmt.Println(returnIndent(index), "Message: ", strList[int(keyWordObj[3].(float64))])
			continue
		}

		fmt.Println(returnIndent(index), "Libname: ", strList[int(keyWordObj[2].(float64))])
		fmt.Println(returnIndent(index), "Name: ", strList[int(keyWordObj[1].(float64))])
		fmt.Println(returnIndent(index), "Args: ", strList[int(keyWordObj[5].(float64))])
		if arr, ok := keyWordObj[9].([]interface{}); ok {
			listKeyWords(arr, strList, index+1)
			// if len(arr) > 0 {
			// 	fmt.Printf("arr: %#v\n", arr)
			// 	fmt.Printf("arr[0]: %#v\n", arr[0])
			// 	if first, ok := arr[0].(float64); ok {
			// 		fmt.Println("First element is a float64:", first)
			// 		listMessage(keyWordObj[9], strList)
			// 		continue
			// 	} else {
			// 		fmt.Println("First element is not a float64")
			// 		listKeyWords(keyWordObj[9], strList)
			// 	}
			// } else {
			// 	fmt.Println("Regular keyword list")
			// 	listKeyWords(keyWordObj[9], strList)
			// }
			// fmt.Println("Next child is empty, abandoning...")
		}
	}
}

// extractSuiteData extracts the JSON data from window.output["suite"] assignment
// and handles variable references like window.sPart0, window.sPart1
func extractSuiteData(htmlContent string) (string, error) {
	// First, extract the main suite assignment
	re := regexp.MustCompile(`window\.output\["suite"\]\s*=\s*(\[.*?\]);`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find window.output[\"suite\"] assignment")
	}
	suiteContent := matches[1]

	// Find all window.sPart variables referenced in the suite
	variablePattern := regexp.MustCompile(`window\.(sPart\d+)`)
	variableMatches := variablePattern.FindAllStringSubmatch(suiteContent, -1)

	// Extract the actual content of each variable
	variableMap := make(map[string]string)
	for _, match := range variableMatches {
		if len(match) >= 2 {
			varName := match[1] // e.g., "sPart0"
			fullVarName := "window." + varName

			// Extract the variable definition
			varPattern := regexp.MustCompile(`window\.` + regexp.QuoteMeta(varName) + `\s*=\s*(\[.*?\]);`)
			varMatches := varPattern.FindStringSubmatch(htmlContent)
			if len(varMatches) >= 2 {
				variableMap[fullVarName] = varMatches[1]
			}
		}
	}

	// Replace variable references with their actual content
	result := suiteContent
	for varRef, varContent := range variableMap {
		result = strings.ReplaceAll(result, varRef, varContent)
	}

	return result, nil
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

	//fmt.Printf("%#v\n", outputArr)

	//fmt.Printf("%#v\n", result)

	arr, ok := result.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}

	// fmt.Printf("%#v\n", arr[5])
	// fmt.Println("Suites")
	// fmt.Printf("%#v\n", arr[6])
	// fmt.Println("Tests")
	// fmt.Printf("%#v\n", arr[7])
	// fmt.Println("Keywords")
	// fmt.Printf("%#v\n", arr[8])
	// fmt.Println("Don't know")
	// fmt.Printf("%#v\n", arr[9])

	listTests(arr[7], outputArr)

	fmt.Println("Suite Keywords:")
	listKeyWords(arr[8], outputArr, 0)

}
