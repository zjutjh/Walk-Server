package main

import (
	"github.com/spf13/cobra"
	"github.com/zjutjh/mygo/foundation/command"
	"github.com/zjutjh/mygo/foundation/httpserver"

	"app/register"
)

func main() {
	command.Execute(
		register.Boot,    // 应用引导注册器
		register.Command, // 应用命令注册器
		// httpserver.CommandRegister(register.Route), // 默认注入HTTP Server注册器
		func(cmd *cobra.Command, args []string) error {
			httpserver.StartHTTPServer(register.Route)
			return nil
		},
	)
}
