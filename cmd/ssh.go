package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"kick/utils"
)

// SSHModel 表示 SSH 配置的状态模型
type SSHModel struct {
	textInput textinput.Model
	err       error
	done      bool
	status    string
	success   bool
}

// NewSSHCmd 创建 ssh 子命令
func NewSSHCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ssh",
		Short: "配置基于 SSH 密钥的认证",
		Long:  `配置 SSH 服务器使用基于密钥的认证并禁用密码登录。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 检查是否为 root 用户
			if !utils.IsRoot() {
				return fmt.Errorf("此命令需要 root 权限，请使用 sudo 运行")
			}

			p := tea.NewProgram(initialModel())
			model, err := p.Run()
			if err != nil {
				return fmt.Errorf("程序运行错误: %v", err)
			}

			finalModel := model.(SSHModel)
			if finalModel.err != nil {
				return finalModel.err
			}

			if !finalModel.success {
				return fmt.Errorf("SSH 配置未完成")
			}

			return nil
		},
	}
}

// 初始化模型
func initialModel() SSHModel {
	ti := textinput.New()
	ti.Placeholder = "ssh-rsa AAAA..."
	ti.Focus()
	ti.CharLimit = 2048
	ti.Width = 80
	ti.Prompt = ""

	return SSHModel{
		textInput: ti,
		err:       nil,
		done:      false,
		status:    "",
		success:   false,
	}
}

// Init 实现 tea.Model 接口的初始化
func (m SSHModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update 实现 tea.Model 接口的更新逻辑
func (m SSHModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// 如果已经完成，只处理退出信息
	if m.done {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc || msg.Type == tea.KeyEnter {
				return m, tea.Quit
			}
		case error:
			m.err = msg
			m.status = "配置过程中发生错误: " + msg.Error()
			return m, tea.Quit
		case sshConfigCompleteMsg:
			m.status = msg.status
			m.success = msg.success
			return m, tea.Quit
		}
		return m, nil
	}

	// 处理文本输入
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if !m.textInput.Focused() {
				return m, tea.Quit
			}

			publicKey := strings.TrimSpace(m.textInput.Value())
			if len(publicKey) > 0 {
				// 验证公钥格式
				if !strings.HasPrefix(publicKey, "ssh-rsa") &&
					!strings.HasPrefix(publicKey, "ssh-ed25519") &&
					!strings.HasPrefix(publicKey, "ecdsa-sha2-") {
					m.err = fmt.Errorf("公钥格式不正确，应该以 ssh-rsa、ssh-ed25519 或 ecdsa-sha2- 开头")
					return m, nil
				}

				m.done = true
				m.status = "正在配置 SSH..."
				return m, tea.Batch(
					performSSHConfig(publicKey),
				)
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View 实现 tea.Model 接口的视图渲染
func (m SSHModel) View() string {
	if m.done {
		return m.status + "\n\n按任意键退出...\n"
	}

	s := "请输入您的公钥 (ssh-rsa 开头的文本):\n\n"
	s += m.textInput.View()
	s += "\n\n按 Enter 继续，Esc 取消\n"

	if m.err != nil {
		s += "\n错误: " + m.err.Error() + "\n"
	}

	return s
}

// sshConfigCompleteMsg 表示 SSH 配置完成的消息
type sshConfigCompleteMsg struct {
	status  string
	success bool
}

// performSSHConfig 执行 SSH 配置的命令
func performSSHConfig(publicKey string) tea.Cmd {
	return func() tea.Msg {
		var status strings.Builder

		status.WriteString("正在修改 SSH 配置...\n")
		err := utils.UpdateSSHConfig()
		if err != nil {
			return fmt.Errorf("修改 SSH 配置失败: %v", err)
		}

		status.WriteString("正在添加公钥到 authorized_keys...\n")
		err = utils.AddPublicKey(publicKey)
		if err != nil {
			return fmt.Errorf("添加公钥失败: %v", err)
		}

		status.WriteString("正在重启 SSH 服务...\n")
		err = utils.RestartSSHService()
		if err != nil {
			return fmt.Errorf("重启 SSH 服务失败: %v", err)
		}

		status.WriteString("\n✅ SSH 配置已成功完成!\n")
		status.WriteString("\n⚠️ 重要提示: 您需要重启系统才能使所有更改生效。\n")
		status.WriteString("在重启前，请确保您可以使用密钥正常登录，以避免被锁定在系统外。\n")

		return sshConfigCompleteMsg{
			status:  status.String(),
			success: true,
		}
	}
}
