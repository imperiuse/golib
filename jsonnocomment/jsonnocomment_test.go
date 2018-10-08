package jsonnocomment

import (
	"fmt"
	"testing"
)

var expectedResult = []byte{123, 10, 32, 32, 34, 110, 97, 109, 101, 34, 58, 32, 34, 116, 101, 115, 116, 32, 74, 83, 79,
	78, 32, 102, 105, 108, 101, 47, 34, 44, 10, 32, 32, 10, 32, 32, 10, 32, 32, 34, 101, 110, 100, 34, 58, 32, 34, 101, 110,
	100, 47, 32, 106, 115, 111, 110, 32, 47, 102, 105, 108, 101, 42, 34, 10, 125}

func TestReadFileAndCleanComment(t *testing.T) {
	bytesResult, err := ReadFileAndCleanComment("./test_json/testFile.json")
	if err != nil || bytesResult == nil {
		t.Errorf("Unexpected error while Read and Clean Json file: %v", err)
		t.Failed()
	}
	if len(bytesResult) != len(expectedResult) {
		t.Errorf("len(result) != len(expectedResult)")
		goto error
	}
	for i := 0; i < len(bytesResult); i++ {
		if bytesResult[i] != expectedResult[i] {
			goto error
		}
	}
	return
error:
	t.Errorf("Unexpected result! Want:%v\nHave:%v\n", expectedResult, bytesResult)
	t.Failed()
}

// Негативный тест проверка на правильную реакцию на несуществующий файл
func TestReadFileAndCleanComment2(t *testing.T) {
	bytesResult, err := ReadFileAndCleanComment("./test_json/noExistFile.json")
	if err == nil || bytesResult != nil {
		t.Errorf("No err while read non exist file: %v", err)
		t.Failed()
	}
}

// Негативный тест проверка на правильную реакцию на плохой файл
func TestReadFileAndCleanComment3(t *testing.T) {
	bytesResult, err := ReadFileAndCleanComment("./test_json/badFile.json")
	if err != nil || bytesResult == nil {
		t.Errorf("Unexpected error while Read and Clean Json file: %v", err)
		t.Failed()
	}
	var expectedResult = `{
  "a": "i'm very very bad file"
}`
	if string(bytesResult) != expectedResult {
		t.Errorf("Unexpected result! Want:%v\nHave:%v\n", string(expectedResult), expectedResult)
		t.Failed()
	}
}

// Негативный тест проверка на правильную реакцию на плохой файл
func TestReadFileAndCleanComment4(t *testing.T) {
	bytesResult, err := ReadFileAndCleanComment("./test_json/badFile2.json")
	if err != nil || bytesResult == nil {
		t.Errorf("Unexpected error while Read and Clean Json file: %v", err)
		t.Failed()
	}
	var expectedResult = `{
  "a": "i'm very very bad file2"
}`
	if string(bytesResult) != expectedResult {
		t.Errorf("Unexpected result! Want:%v\nHave:%v\n", string(expectedResult), expectedResult)
		t.Failed()
	}
}

// Простой пример использования
func ExampleReadFileAndCleanComment() {
	// file: testFile.json:
	//`{
	//  "name": "test */JSON file/",
	//  //"comment_line": "not /see this line of //text /*comment too */",/
	//  /*"comment_block": "not see *this* block of //text",*/
	//  "end": "end/ json /file*"
	//}//`

	bytes, err := ReadFileAndCleanComment("./test_json/testFile.json")
	if err != nil {
		print("Unexpected error while Read and Clean Json file: %v", err)
	}
	fmt.Println(string(bytes)) // Output: {
	//	"name": "test JSON file/",
	//
	//
	//		"end": "end/ json /file*"
	//}

}
