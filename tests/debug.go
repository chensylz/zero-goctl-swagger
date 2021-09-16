package main

import (
	"fmt"
	"github.com/1278651995/goctl-swagger/generate"
	"github.com/tal-tech/go-zero/tools/goctl/api/parser"
	plugin2 "github.com/tal-tech/go-zero/tools/goctl/plugin"
)

const userAPI = "./tests/user.api"

func main() {
	result, err := parser.Parse(userAPI)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &plugin2.Plugin{
		Api:         result,
		ApiFilePath: userAPI,
		Style:       "",
		Dir:         ".",
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	err = generate.Do("./tests/user.api", "", "/", p)
	if err != nil {
		fmt.Println(err)
	}
}
