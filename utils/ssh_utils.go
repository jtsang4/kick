package utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// IsRoot 检查当前程序是否以 root 权限运行
func IsRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}

// UpdateSSHConfig 更新 SSH 配置文件
func UpdateSSHConfig() error {
	sshConfigPath := "/etc/ssh/sshd_config"

	// 读取当前配置
	content, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return err
	}

	configStr := string(content)
	lines := strings.Split(configStr, "\n")

	// 需要修改的配置项
	configUpdates := map[string]string{
		"PubkeyAuthentication":   "yes",
		"AuthorizedKeysFile":     ".ssh/authorized_keys .ssh/authorized_keys2",
		"PasswordAuthentication": "no",
		"PermitRootLogin":        "prohibit-password",
		"ClientAliveInterval":    "60",
		"ClientAliveCountMax":    "10",
	}

	// 检查并更新配置
	updatedConfig := make(map[string]bool)
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		for key, value := range configUpdates {
			if strings.HasPrefix(trimmedLine, key+" ") || strings.HasPrefix(trimmedLine, key+"\t") {
				lines[i] = key + " " + value
				updatedConfig[key] = true
				break
			}
		}
	}

	// 添加缺失的配置项
	for key, value := range configUpdates {
		if !updatedConfig[key] {
			lines = append(lines, key+" "+value)
		}
	}

	// 检查 sshd_config.d 目录，避免被其中的配置覆盖
	configDirPath := "/etc/ssh/sshd_config.d"
	if _, err := os.Stat(configDirPath); !os.IsNotExist(err) {
		files, err := os.ReadDir(configDirPath)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".conf") {
					filePath := filepath.Join(configDirPath, file.Name())
					fileContent, err := os.ReadFile(filePath)
					if err == nil {
						// 检查文件中是否有密码登录的配置
						if strings.Contains(string(fileContent), "PasswordAuthentication yes") {
							// 如果发现允许密码登录的配置，添加注释
							newContent := strings.Replace(
								string(fileContent),
								"PasswordAuthentication yes",
								"#PasswordAuthentication yes # 被 kick 工具禁用",
								-1,
							)
							if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
								return fmt.Errorf("无法修改配置文件 %s: %v", filePath, err)
							}
						}
					}
				}
			}
		}
	}

	// 写回配置文件
	return os.WriteFile(sshConfigPath, []byte(strings.Join(lines, "\n")), 0644)
}

// AddPublicKey 将公钥添加到 authorized_keys 文件
func AddPublicKey(publicKey string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sshDir := homeDir + "/.ssh"

	// 确保 .ssh 目录存在
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		err = os.Mkdir(sshDir, 0700)
		if err != nil {
			return err
		}
	}

	authKeysPath := sshDir + "/authorized_keys"

	// 检查是否已存在该公钥
	existingKeys := ""
	if _, err := os.Stat(authKeysPath); !os.IsNotExist(err) {
		content, err := os.ReadFile(authKeysPath)
		if err != nil {
			return err
		}
		existingKeys = string(content)
	}

	// 如果公钥已存在，则不重复添加
	if strings.Contains(existingKeys, publicKey) {
		return nil
	}

	// 追加公钥
	if existingKeys != "" && !strings.HasSuffix(existingKeys, "\n") {
		existingKeys += "\n"
	}
	existingKeys += publicKey + "\n"

	// 写入公钥文件并设置权限
	err = os.WriteFile(authKeysPath, []byte(existingKeys), 0600)
	if err != nil {
		return err
	}

	return os.Chmod(authKeysPath, 0600)
}

// RestartSSHService 重启 SSH 服务
func RestartSSHService() error {
	// 尝试不同的重启命令，兼容不同的系统
	restartCommands := []string{
		"service ssh restart",
		"systemctl restart ssh",
		"systemctl restart sshd",
		"service sshd restart",
	}

	var lastErr error
	for _, cmd := range restartCommands {
		parts := strings.Split(cmd, " ")
		if err := exec.Command(parts[0], parts[1:]...).Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("无法重启 SSH 服务: %v", lastErr)
	}
	return fmt.Errorf("无法重启 SSH 服务，请手动重启")
}
