package main

import "C"

import (
	"fmt"

	"github.com/aws-cloudformation/rain/cfn/format"
	"github.com/aws-cloudformation/rain/cfn/parse"
)

func main() {
	fmt.Println("Make it rain!")
}

//export ToJson
func ToJson(input *C.char) *C.char {
	t, err := parse.String(C.GoString(input))
	if err != nil {
		panic(err)
	}

	return C.CString(format.Template(t, format.Options{
		Style: format.JSON,
	}))
}
