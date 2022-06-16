package action

import (
	"github.com/1278651995/zero-goctl-swagger/generate"
	"github.com/urfave/cli/v2"
	plugin2 "github.com/zeromicro/go-zero/tools/goctl/plugin"
)

func Generator(ctx *cli.Context) error {
	fileName := ctx.String("filename")

	if len(fileName) == 0 {
		fileName = "rest.swagger.json"
	}

	p, err := plugin2.NewPlugin()
	if err != nil {
		return err
	}
	basepath := ctx.String("basepath")
	host := ctx.String("host")
	return generate.Do(fileName, host, basepath, p)
}
