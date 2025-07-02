package file

import (
	"crypto/md5"
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/nginxpath"
	"io"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
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
	if err := handleConfigUpdate(path, content); err != nil {
		return err
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

// 处理配置文件更新
func handleConfigUpdate(targetPath string, newContent []byte) error {
	tmpPath := targetPath + ".tmp"

	// 如果存在，先比对
	_, err := os.Stat(targetPath)
	if err == nil {
		oldMD5, err := getFileMD5(targetPath)
		if err != nil {
			return fmt.Errorf("failed to calculate MD5 of the original file: %v", err)
		}

		newMD5 := getContentMD5(newContent)

		if oldMD5 == newMD5 {
			fmt.Println("[INFO] content MD5 consistent, no need to update")
			return nil
		}
	}

	// 写临时文件
	if err := writeToFile(tmpPath, newContent); err != nil {
		return fmt.Errorf("failed to write temporary file: %v", err)
	}

	// 检测
	if err := checkNginxConfig(); err != nil {
		fmt.Println("[WARN] nginx -t test failed，rollback")
		if err := os.Remove(tmpPath); err != nil {
			klog.Warning(fmt.Sprintf("[WARN] nginx -t test failed，rollback"))
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

	fmt.Printf("[SUCCESS] update %s completed and reload\n", targetPath)
	return nil
}
