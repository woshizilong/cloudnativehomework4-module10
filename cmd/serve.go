package cmd

import (
	"github.com/kabacloud/cloudnativehomework4-module10/service"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "服务",
	Long: `本服务可运行在kubernetes集群下。提供如下服务：
	- 存活探针 : /healthz
	- 就绪探针 : /readyz
	- 监控指标 : /metrics
	- 打印服务信息 : /info`,
	Run: func(cmd *cobra.Command, args []string) {
		execServe(args)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func execServe(args []string) {
	_ = service.Start(MainContext)
}
