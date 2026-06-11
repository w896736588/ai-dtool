package p_db

import (
	"strings"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsssh"
	"github.com/w896736588/go-tool/gstool"
)

func buildConfiguredSshConfig(sshConfig map[string]any) *gsssh.SshConfig {
	cfg := &gsssh.SshConfig{
		Name:                 cast.ToString(sshConfig[`name`]),
		Host:                 cast.ToString(sshConfig[`host`]),
		Port:                 cast.ToString(sshConfig[`port`]),
		UserName:             cast.ToString(sshConfig[`username`]),
		Password:             cast.ToString(sshConfig[`password`]),
		ConnectTimeoutSecond: cast.ToInt(sshConfig[`connect_timeout`]),
		ExecAfterConnect:     strings.TrimSpace(cast.ToString(sshConfig[`post_connect_cmds`])),
	}
	gstool.FmtPrintlnLogTime(
		`[p_db.NewConfiguredSshBridge] build ssh config name=%s host=%s port=%s timeout=%d exec_after_connect=%q`,
		cfg.Name, cfg.Host, cfg.Port, cfg.ConnectTimeoutSecond, cfg.ExecAfterConnect,
	)
	return cfg
}

// NewConfiguredSshBridge reuses SSH settings for DB/Redis tunnels, including jumpserver commands.
func NewConfiguredSshBridge(sshConfig map[string]any) *gsssh.SshBridge {
	return gsssh.NewSshBridge(gsssh.NewSsh(buildConfiguredSshConfig(sshConfig)))
}
