package jsonnocomment

import (
	"fmt"
	"testing"
)

func TestReadFileAndCleanComment(t *testing.T) {
	bytes, err := ReadFileAndCleanComment("./testFile.json")
	if err != nil {
		t.Errorf("Unexpected error while Read and Clean Json file: %v", err)
		t.Failed()
	}
	if string(bytes) != `{  "name": "test JSON file",      "end": "end json file"}`{
		t.Errorf("Unexpected result! Want:%v\nHave:%v\n",
			`{  "name": "test JSON file",      "end": "end json file"}`,string(bytes))
		t.Failed()
	}
}

// Простой пример использования
func ExampleReadFileAndCleanComment() {
	// file: testFile.json:
	//`{
  	//	"name": "test JSON file",
  	//	//"comment_line": "not see this line of //text /*comment too */",
  	//	/*"comment_block": "not see this block of //text",*/
 	//	"end": "end json file"
	//}`

	bytes, err := ReadFileAndCleanComment("./testFile.json")
	if err != nil {
		print("Unexpected error while Read and Clean Json file: %v", err)
	}
	fmt.Println(string(bytes)) // Output: {  "name": "test JSON file",      "end": "end json file"}
}
