package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const INDENT = 4

var baseMillis int64

func returnIndent(level int) string {
	return strings.Repeat(" ", level*INDENT)
}

func listSuites(elemList interface{}, strList []string) {
	suitesArr, ok := elemList.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}
	fmt.Println("Nr. of suites:", len(suitesArr))

	for _, val := range suitesArr {
		//fmt.Printf("Tests Index: %d, Value: %#v\n", i, val)
		suitesObj, ok := val.([]interface{})
		if !ok {
			fmt.Println("Unexpected type")
			return
		}

		fmt.Println("Suite name: ", strList[int(suitesObj[0].(float64))])
		fmt.Println("Suite doc: ", strList[int(suitesObj[3].(float64))])
		fmt.Println("Suite source: ", strList[int(suitesObj[1].(float64))])
		fmt.Println("Suite relative source: ", strList[int(suitesObj[2].(float64))])

		// suite can have other suites and tests
		listSuites(suitesObj[6], strList)
		listTests(suitesObj[7], strList)
		listKeyWords(suitesObj[8], strList, 0)

	}
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

		fmt.Println("Test name: ", strList[int(testObj[0].(float64))])
		fmt.Println("Doc: ", strList[int(testObj[2].(float64))])

		listKeyWords(testObj[5], strList, 0)
	}
}

func listKeyWords(elemList interface{}, strList []string, index int) {
	keyWordArr, ok := elemList.([]interface{})
	if !ok {
		fmt.Println("Unexpected type, not an array")
		return
	}
	//fmt.Println("Nr. of keywords:", len(keyWordArr))

	for _, val := range keyWordArr {
		//fmt.Printf("Keywords Index: %d, Value: %#v\n", i, val)
		keyWordObj, ok := val.([]interface{})
		if !ok {
			fmt.Println("Unexpected type in listKeyWords")
			return
		}

		if len(keyWordObj) < 5 {
			//this is a message, not a keyword
			if strList[int(keyWordObj[3].(float64))][0] == '*' {
				fmt.Println(returnIndent(index), "Message: ", strList[int(keyWordObj[3].(float64))][1:])
			}
			//else it's zipped message, TBD later
			continue
		}

		fmt.Println(returnIndent(index), "Name: ", strList[int(keyWordObj[2].(float64))][1:]+"."+strList[int(keyWordObj[1].(float64))][1:])
		//fmt.Println(returnIndent(index), "Name: ", strList[int(keyWordObj[1].(float64))])
		fmt.Println(returnIndent(index), "Args: ", strList[int(keyWordObj[5].(float64))][1:])

		//fmt.Println(returnIndent(index), "Time: ", strList[int(keyWordObj[8].(float64))][1:])
		if times, ok := keyWordObj[8].([]interface{}); ok {
			startMillis := int64(times[1].(float64))
			elapsedMillis := int64(times[2].(float64))

			// Convert start time from baseMillis epoch to actual timestamp
			actualStartMillis := baseMillis + startMillis
			startTime := time.Unix(actualStartMillis/1000, (actualStartMillis%1000)*1000000)

			// Calculate end time by adding elapsed time
			actualEndMillis := actualStartMillis + elapsedMillis
			endTime := time.Unix(actualEndMillis/1000, (actualEndMillis%1000)*1000000)

			fmt.Printf("%s Start: %s End: %s Elapsed: %dms)\n",
				returnIndent(index),
				startTime.Format("2006-01-02 15:04:05.000"),
				endTime.Format("2006-01-02 15:04:05.000"),
				elapsedMillis)
			//fmt.Printf("%sEnd:   %s (elapsed: %dms)\n", returnIndent(index), endTime.Format("2006-01-02 15:04:05.000"), elapsedMillis)
		} else {
			fmt.Println("Unexpected type in listKeyWords - times")
			return
		}

		if arr, ok := keyWordObj[9].([]interface{}); ok {
			listKeyWords(arr, strList, index+1)
		} else {
			fmt.Println("Unexpected type moving to next keyword, not an array")
			return
		}
	}
}

// extractSuiteData extracts the JSON data from window.output["suite"] assignment
// and handles nested variable references like window.sPart0, window.sPart1, etc.
func extractSuiteData(htmlContent string) (string, error) {
	// First, extract the main suite assignment
	re := regexp.MustCompile(`window\.output\["suite"\]\s*=\s*(\[.*?\]);`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find window.output[\"suite\"] assignment")
	}
	suiteContent := matches[1]

	// Extract all variable definitions from HTML
	allVariables := make(map[string]string)
	variablePattern := regexp.MustCompile(`window\.(sPart\d+)\s*=\s*(\[.*?\]);`)
	varMatches := variablePattern.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range varMatches {
		if len(match) >= 3 {
			varName := "window." + match[1] // e.g., "window.sPart0"
			varContent := match[2]          // the actual array content
			allVariables[varName] = varContent
		}
	}

	// Recursively resolve all variable references
	resolved := make(map[string]string)
	result := resolveVariableReferences(suiteContent, allVariables, resolved)

	return result, nil
}

// resolveVariableReferences recursively resolves nested variable references
func resolveVariableReferences(content string, allVariables map[string]string, resolved map[string]string) string {
	// Find all variable references in current content
	// Match complete variable name followed by delimiter to avoid partial matches like sPart1 vs sPart10
	variableRefPattern := regexp.MustCompile(`(window\.sPart\d+)([,\]])`)
	refMatches := variableRefPattern.FindAllStringSubmatch(content, -1)

	result := content
	for _, match := range refMatches {
		if len(match) >= 3 {
			fullVarName := match[1] // e.g., "window.sPart0" (without delimiter)
			delimiter := match[2]   // the delimiter: ',' or ']'
			fullMatch := match[0]   // complete match including delimiter

			// Skip if already resolved
			if resolvedContent, exists := resolved[fullVarName]; exists {
				// Replace the exact match (variable + delimiter) with (resolved content + delimiter)
				result = strings.ReplaceAll(result, fullMatch, resolvedContent+delimiter)
				continue
			}

			// Get the variable definition
			if varContent, exists := allVariables[fullVarName]; exists {
				// Recursively resolve any variables within this variable's content
				resolvedVarContent := resolveVariableReferences(varContent, allVariables, resolved)

				// Cache the resolved content
				resolved[fullVarName] = resolvedVarContent

				// Replace the exact match (variable + delimiter) with (resolved content + delimiter)
				result = strings.ReplaceAll(result, fullMatch, resolvedVarContent+delimiter)
			}
		}
	}

	return result
}

// extractStringsData extracts and concatenates all JSON data from multiple window.output["strings"] concat assignments
func extractStringsData(htmlContent string) (string, error) {
	re := regexp.MustCompile(`window\.output\["strings"\].*?\.concat\((\[.*?\])\);`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("could not find any window.output[\"strings\"] concat assignments")
	}

	// Collect all string arrays from the concatenations
	var allArrays []string
	for _, match := range matches {
		if len(match) >= 2 {
			allArrays = append(allArrays, match[1])
		}
	}

	// Parse each array and merge them into one final array
	var finalStrings []string

	for _, arrayStr := range allArrays {
		var tempArray []string
		cleaned := strings.ReplaceAll(arrayStr, `\x3c`, "<")
		err := json.Unmarshal([]byte(cleaned), &tempArray)
		if err != nil {
			return "", fmt.Errorf("error parsing string array: %v", err)
		}
		finalStrings = append(finalStrings, tempArray...)
	}

	// Convert back to JSON string
	finalJSON, err := json.Marshal(finalStrings)
	if err != nil {
		return "", fmt.Errorf("error creating final JSON: %v", err)
	}

	return string(finalJSON), nil
}

// extractBaseMillis extracts the baseMillis value from window.output["baseMillis"] assignment
func extractBaseMillis(htmlContent string) error {
	re := regexp.MustCompile(`window\.output\["baseMillis"\]\s*=\s*(\d+);`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return fmt.Errorf("could not find window.output[\"baseMillis\"] assignment")
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing baseMillis value: %v", err)
	}

	baseMillis = value
	return nil
}

// readHTMLFile reads the HTML file and extracts the required JSON data
func readHTMLFile(filename string) (string, string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("error reading file: %v", err)
	}

	htmlContent := string(content)

	// Extract baseMillis value and store in package variable
	err = extractBaseMillis(htmlContent)
	if err != nil {
		return "", "", err
	}

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

	//cleaned := strings.ReplaceAll(output_strings, `\x3c`, "<")

	var result interface{}
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var outputArr []string
	err = json.Unmarshal([]byte(output_strings), &outputArr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	arr, ok := result.([]interface{})
	if !ok {
		fmt.Println("Unexpected type")
		return
	}

	//listSuites(arr, outputArr)

	fmt.Printf("BaseMillis: %d\n", baseMillis)

	listSuites(arr[6], outputArr)
	listTests(arr[7], outputArr)
	listKeyWords(arr[8], outputArr, 0)
}
