package main

import (
	"encoding/json"
	"fmt"
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

func main() {
	data := `[1,2,3,0,[],[1,0,16],[],[[4,0,5,[],[1,14,1],[[0,6,7,0,8,9,0,0,[1,15,0],[]]]],[10,0,0,[],[1,15,1],[[0,6,7,0,8,11,0,0,[1,15,0],[]],[0,12,7,0,13,14,0,0,[1,16,0],[[16,2,14]]]]]],[[1,15,0,0,0,16,0,0,[1,14,1],[[0,6,7,0,8,17,0,0,[1,14,0],[]]]]],[2,2,0,0]]`
	output_strings := `["*","*Hello","*/home/fkovacs/robot/hello.robot","*hello.robot","*My Hello","*<p>Here is the placeholder for doc\x3c/p>","*Log To Console","*BuiltIn","*<p>Logs the given message to the console.\x3c/p>","*Hello world","*Another TC","*Started the next","*Log","*<p>Logs the given message with the given level.\x3c/p>","*This is logging as well","*My Log","*Wow","*${text}"]`

	cleaned := strings.ReplaceAll(output_strings, `\x3c`, "<")

	var result interface{}
	err := json.Unmarshal([]byte(data), &result)
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
