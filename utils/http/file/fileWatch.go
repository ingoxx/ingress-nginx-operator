package file

import (
	"crypto/md5"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/nginxpath"
	"io"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"
)

func StartWatch() error {
	// 创建新的 watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Error("create watcher failed:", err)
		return err
	}
	defer watcher.Close()

	done := make(chan bool)

	// 启动监听逻辑
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// 打印事件类型
				klog.Info(fmt.Sprintf("[event] %s -> %s\n", event.Op, event.Name))

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Error("[error]", err)
			}
		}
	}()

	// 监控 nginx.conf 文件
	nginxConfPath := nginxpath.NginxMainConf
	err = watcher.Add(nginxConfPath)
	if err != nil {
		klog.Error(fmt.Sprintf("failed to monitor '%s', error '%v'", nginxConfPath, err))
		return err
	}
	klog.Info("listening:", nginxConfPath)

	// 监控 conf.d 目录及其文件
	err = watcher.Add(nginxpath.NginxConfDir)
	if err != nil {
		klog.Error(fmt.Sprintf("failed to monitor '%s' dir, error '%s'", nginxpath.NginxConfDir, err))
		return err
	}
	klog.Info("listening dir:", nginxpath.NginxConfDir)

	// 同时监听当前目录下已存在的文件
	err = addExistingFiles(watcher, nginxpath.NginxConfDir)
	if err != nil {
		klog.Error(fmt.Sprintf("failed to listen to files in '%s' directory, error '%v'", nginxpath.NginxConfDir, err))
		return err
	}

	<-done

	return nil
}

// addExistingFiles 会将目录下现有文件添加到 watcher 中
func addExistingFiles(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				klog.Error(fmt.Sprintf("failed to listen to files: %s (%v)\n", path, err))
			} else {
				klog.Info("listening file:", path)
			}
		}
		return nil
	})
}

// SaveToFile 将 content 内容写入到 filepath 指定的文件，自动创建目录。
func SaveToFile(path string, content []byte) error {
	dir := filepath.Dir(path)

	// 自动创建目录（如果不存在）
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件（覆盖写入）
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// 计算文件 MD5
func getFileMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// 计算内容 MD5
func getContentMD5(content []byte) string {
	sum := md5.Sum(content)
	return fmt.Sprintf("%x", sum)
}

// 写文件
func writeToFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// nginx -t
func checkNginxConfig() error {
	cmd := exec.Command("nginx", "-t")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// nginx reload
func reloadNginx() error {
	cmd := exec.Command("nginx", "-s", "reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// 检查 nginx 是否存在
func isNginxRunning() bool {
	cmd := exec.Command("pgrep", "nginx")
	out, _ := cmd.Output()
	return len(out) > 0
}

// 启动 nginx
func startNginx() error {
	cmd := exec.Command("nginx", "-c", nginxpath.NginxMainConf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 使用 Start() 而不是 Run()，这样不会阻塞
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

// IsNginxRunning 检查 nginx 是否在跑
func IsNginxRunning() error {
	if !isNginxRunning() {
		klog.Info("[INFO] nginx not running, try starting......")
		if err := startNginx(); err != nil {
			return fmt.Errorf("attempt to start nginx failed: %v", err)
		}
		if !isNginxRunning() {
			return fmt.Errorf("nginx still cannot detect processes after startup")
		}
		klog.Info("[INFO] nginx started successfully")
	} else {
		klog.Info("[INFO] nginx started successfully")
	}

	return nil
}

// HandleConfigUpdate 处理配置文件更新
func HandleConfigUpdate(targetPath string, newContent []byte) error {
	tmpPath := targetPath + ".tmp"

	// 如果存在，先比对
	_, err := os.Stat(targetPath)
	if err == nil {
		if len(newContent) == 0 {
			if err := os.Remove(targetPath); err != nil {
				return err
			}
			if err := reloadNginx(); err != nil {
				return fmt.Errorf("failed to nginx reload: %v", err)
			}
			return nil
		}

		oldMD5, err := getFileMD5(targetPath)
		if err != nil {
			return fmt.Errorf("failed to calculate MD5 of the original file: %v", err)
		}

		newMD5 := getContentMD5(newContent)
		if oldMD5 == newMD5 {
			klog.Info("[INFO] content MD5 consistent, no need to update")
			return nil
		}
	}

	// 写临时文件
	if err := writeToFile(tmpPath, newContent); err != nil {
		return fmt.Errorf("failed to write temporary file: %v", err)
	}

	// 检测
	if err := checkNginxConfig(); err != nil {
		if err := os.Remove(tmpPath); err != nil {
			klog.Warning(fmt.Sprintf("[WARN] nginx -t test failed, rollback"))
		}
		return err
	}

	// OK，替换
	if err := writeToFile(targetPath, newContent); err != nil {
		return fmt.Errorf("failed to overwrite official documents: %v", err)
	}

	// reload
	if err := reloadNginx(); err != nil {
		return fmt.Errorf("failed to nginx reload: %v", err)
	}

	if err := os.Remove(tmpPath); err != nil {
		return err
	}

	klog.Infof("[SUCCESS] update %s completed and reload\n", targetPath)
	return nil
}

func HandleDeleteNgxConfig(targetFile string) error {
	if _, err := os.Stat(targetFile); err != nil {
		return nil
	}

	if err := os.Remove(targetFile); err != nil {
		return err
	}

	// reload
	if err := reloadNginx(); err != nil {
		return fmt.Errorf("failed to nginx reload: %v", err)
	}

	return nil
}
