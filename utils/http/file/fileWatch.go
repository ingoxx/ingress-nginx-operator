package file

import (
	"fmt"
	"github.com/ingoxx/ingress-nginx-operator/utils/http/nginxpath"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func StartWatch() error {
	// 创建新的 watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Error("创建 watcher 失败:", err)
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
				klog.Info(fmt.Sprintf("[事件] %s -> %s\n", event.Op, event.Name))

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Error("[错误]", err)
			}
		}
	}()

	// 监控 nginx.conf 文件
	nginxConfPath := nginxpath.NginxMainConf
	err = watcher.Add(nginxConfPath)
	if err != nil {
		klog.Error("无法监控 nginx.conf:", err)
		return err
	}
	klog.Info("已监听:", nginxConfPath)

	// 监控 conf.d 目录及其文件
	confDir := nginxpath.NginxConfDir
	err = watcher.Add(confDir)
	if err != nil {
		klog.Error("无法监控 conf.d 目录:", err)
		return err
	}
	klog.Info("已监听目录:", confDir)

	// 同时监听当前目录下已存在的文件
	err = addExistingFiles(watcher, confDir)
	if err != nil {
		klog.Error("添加 conf.d 下文件失败:", err)
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
				klog.Error(fmt.Sprintf("监听文件失败: %s (%v)\n", path, err))
			} else {
				klog.Info("已监听文件:", path)
			}
		}
		return nil
	})
}
