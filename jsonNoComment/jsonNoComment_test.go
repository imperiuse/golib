package jsonNoComment

import (
	"fmt"
)

func ExampleReadFileAndCleanComment() {
	bytes, err := ReadFileAndCleanComment("./testFile.json")
	if err != nil {
		print("Unexpected error while Read and Clean Json file: %v", err)
	}
	fmt.Println(string(bytes)) // Output: {  "name": "test JSON file",      "end": "end json file"}

}
