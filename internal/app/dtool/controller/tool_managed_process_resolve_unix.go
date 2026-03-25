//go:build !windows

package controller

func resolveManagedProcessStartConfig(config managedProcessConfig) (managedProcessConfig, error) {
	return config, nil
}
