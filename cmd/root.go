package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd 创建根命令
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kick",
		Short: "Kick 是一个自动化服务器任务的 CLI 工具",
		Long: `Kick 帮助自动化常见的服务器配置任务，包括 SSH 设置。
使用 'kick ssh' 命令来配置 SSH 密钥认证。`,
	}

	// 添加 ssh 子命令
	rootCmd.AddCommand(NewSSHCmd())

	return rootCmd
}
