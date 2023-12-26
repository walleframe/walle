package disk_monitor

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

// DiskNotifier 磁盘监控变动接口
type DiskNotifier interface {
	Prepare(logger *zaplog.Logger, fileName string)
	Change(logger *zaplog.Logger, fileName string, data []byte) error
	Finish(logger *zaplog.Logger, fileName string)
}

// DiskNotifierChangeFunc 监控文件变化函数
type DiskNotifierChangeFunc func(logger *zaplog.Logger, fileName string, data []byte) error

func (DiskNotifierChangeFunc) Prepare(logger *zaplog.Logger, fileName string) {
	return
}
func (df DiskNotifierChangeFunc) Change(logger *zaplog.Logger, fileName string, data []byte) error {
	return df(logger, fileName, data)
}

func (DiskNotifierChangeFunc) Finish(logger *zaplog.Logger, fileName string) {
	return
}

var _ DiskNotifier = DiskNotifierChangeFunc(nil)

type monitorFuncs struct {
	prepare func(logger *zaplog.Logger, fileName string)
	change  func(logger *zaplog.Logger, fileName string, data []byte) error
	finish  func(logger *zaplog.Logger, fileName string)
}

func (mf *monitorFuncs) Prepare(logger *zaplog.Logger, fileName string) {
	if mf.prepare != nil {
		mf.prepare(logger, fileName)
	}
	return
}

func (mf *monitorFuncs) Change(logger *zaplog.Logger, fileName string, data []byte) error {
	if mf.change != nil {
		return mf.change(logger, fileName, data)
	}
	return nil
}

func (mf *monitorFuncs) Finish(logger *zaplog.Logger, fileName string) {
	if mf.finish != nil {
		mf.finish(logger, fileName)
	}
	return
}

var _ DiskNotifier = (*monitorFuncs)(nil)

func DiskNoitifierFunc(prepare func(logger *zaplog.Logger, fileName string),
	change func(logger *zaplog.Logger, fileName string, data []byte) error,
	finish func(logger *zaplog.Logger, fileName string)) DiskNotifier {
	return &monitorFuncs{
		prepare: prepare,
		change:  change,
		finish:  finish,
	}
}

// MonitorOption use for process
//
//go:generate gogen option -n MonitorOption -o options.go
func walleMonitorOption() interface{} {
	return map[string]interface{}{
		// log interface
		"Logger": (*zaplog.Logger)(zaplog.GetFrameLogger()),
		// file ext
		"FileExt": "*",
		// monitor file or paths
		"Paths": []string{},
		// retry time limit, 0:disable retry
		"RetryLimit": 3,
		// retry interval
		"RetryInterval": time.Duration(time.Second * 10),
		// monitor notify interface
		"Notifier": DiskNotifier(nil),
		// app stopper
		"Stoper": app.Stoper(nil),
	}
}

type retryFile struct {
	file   string // 通知失败文件
	md5sum string // 通知失败时文件md5
	times  int    // 失败次数
}

type DiskMonitor struct {
	// 底层文件监听接口
	watcher *fsnotify.Watcher

	// 文件缓存信息
	filesMd5 map[string]string // 文件md5缓存 map[file]md5
	md5Lock  sync.Mutex        // 文件版本锁
	paths    sync.Map          // 路径过滤 - 目录监控
	files    sync.Map          // 文件过滤 - 文件监控
	fileDirs map[string]string // 文件路径 - 文件监控
	fileLock sync.Mutex        // 路径锁   - 文件监控

	// 通知失败重试队列
	retry chan *retryFile
	// 选项配置
	opts *MonitorOptions

	// 状态
	init   bool // 是否异步通知
	ctx    context.Context
	cancel func()
}

func NewMonitor(opt ...MonitorOption) *DiskMonitor {
	opts := NewMonitorOptions(opt...)

	if strings.Contains(opts.FileExt, "*") {
		opts.FileExt = "*"
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DiskMonitor{
		init:     true,
		filesMd5: make(map[string]string),
		fileDirs: make(map[string]string),
		retry:    make(chan *retryFile, 512),
		opts:     opts,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (dm *DiskMonitor) Start() (err error) {
	notifyer := dm.opts.Notifier
	logger := dm.opts.Logger

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	dm.watcher = watcher

	notifyer.Prepare(logger, "")
	for _, v := range dm.opts.Paths {
		err = dm.Monitor(v)
		if err != nil {
			return
		}
	}
	notifyer.Finish(logger, "")
	// 初始化结束
	dm.init = false

	// check app stoper
	var ch <-chan struct{}
	if dm.opts.Stoper == nil {
		ch = make(chan struct{})
	} else {
		ch = dm.opts.Stoper.GetStopChan()
	}
	// 监控文件变化
	go dm.monitorChange(dm.ctx, ch)

	return
}

func (dm *DiskMonitor) Stop() {
	if dm == nil {
		return
	}
	if dm.cancel != nil {
		dm.cancel()
	}
	if dm.watcher != nil {
		dm.watcher.Close()
	}
}

// Monitor 监控文件或者目录变化
func (dm *DiskMonitor) Monitor(fileOrDir string) error {
	log := dm.opts.Logger.New("DiskMonitor.Monitor")
	if !fileExists(fileOrDir) {
		log.Error("file or dir not exists", zap.String("fileOrDir", fileOrDir))
		return fmt.Errorf("path or file (%s) not exists", fileOrDir)
	}
	absValue, err := filepath.Abs(fileOrDir)
	if err != nil {
		log.Error("file or dir convert abs failed", zap.String("fileOrDir", fileOrDir), zap.String("abs", absValue))
		return err
	}
	if fileIsDir(absValue) {
		return dm.monitorPath(absValue, true)
	}
	return dm.monitorFile(absValue)
}

// 是否监控了目录
func (dm *DiskMonitor) isMonitorPath(path string) (m bool) {
	dm.paths.Range(func(k, v interface{}) bool {
		mp := k.(string)
		if hasPrefix(mp, path) {
			m = true
			return false
		}
		return true
	})
	return
}

// 是否监控了文件
func (dm *DiskMonitor) isMonitorFile(file string) bool {
	dir := filepath.Dir(file)
	if dm.isMonitorPath(dir) {
		return true
	}
	if _, ok := dm.files.Load(file); ok {
		return true
	}
	return false
}

// 读取文件,检测并通知变动
func (dm *DiskMonitor) loadFileAndNotify(file, check string) error {
	log := dm.opts.Logger.New("DiskMonitor.loadFileAndNotify")
	// 文件过滤判定
	if dm.opts.FileExt != "*" && dm.opts.FileExt != filepath.Ext(file) {
		log.Info("ignore file.", zap.String("file", file), zap.String("ext", dm.opts.FileExt))
		return nil
	}
	// 读取文件
	data, err := os.ReadFile(file)
	if err != nil {
		log.Error("read file failed.", zap.String("file", file))
		return err
	}

	// md5计算
	md5sum := fmt.Sprintf("%x", md5.Sum(data))
	// 检查 - 如果之前通知错误的文件md5值变了.不进行重试.(比如短时间内连续两次更新了文件)
	if check != "" {
		if check != md5sum {
			// 忽略变更的重试文件
			log.Warn("recv retry file ignore. file change.", zap.String("file", file),
				zap.String("check", check), zap.String("cur_md5", md5sum))
			return nil
		}
	} else {
		// 不是发生错误重试的,读取缓存,比较文件md5
		dm.md5Lock.Lock()
		curMd5, ok := dm.filesMd5[file]
		if ok {
			if curMd5 == md5sum {
				dm.md5Lock.Unlock()
				log.Warn("file not modify", zap.String("file", file), zap.String("last_md5", curMd5), zap.String("cur_md5", md5sum))
				return nil
			}
		}
		dm.filesMd5[file] = md5sum
		dm.md5Lock.Unlock()
		log.Info("file modify", zap.String("file", file), zap.String("last_md5", curMd5), zap.String("cur_md5", md5sum))
	}

	// 通知函数
	notify := func() (err error) {
		err = dm.opts.Notifier.Change(dm.opts.Logger, file, data)
		if err != nil {
			log.Error("notify file change failed. ", zap.String("file", file), zap.String("md5", md5sum))
			if !dm.init && check == "" && dm.opts.RetryLimit > 0 {
				// 放入重试队列
				dm.retry <- &retryFile{
					file:   file,
					md5sum: md5sum,
				}
			}
		}
		return err
	}
	// 进行处理
	if dm.init {
		return notify()
	}
	go notify()
	return nil
}

// 监控文件
func (dm *DiskMonitor) monitorFile(file string) (err error) {
	log := dm.opts.Logger.New("DiskMonitor.monitorFile")
	err = dm.loadFileAndNotify(file, "")
	if err != nil {
		log.Error("watch file failed.", zap.String("file", file), zap.Error(err))
		return
	}

	// 监控文件
	err = dm.watcher.Add(file)
	if err != nil {
		log.Error("watch file failed.", zap.String("file", file), zap.Error(err))
	}
	log.Info("monitor file.", zap.String("file", file))
	path := filepath.Dir(file)

	// 递归获取连接目录的上一级目录
	father := path
	var symlik string
	for {
		symlik, err = filepath.EvalSymlinks(father)
		if err != nil {
			log.Error("get real path failed.", zap.String("src", path), zap.String("father", father), zap.Error(err))
		}
		if symlik == father {
			break
		}
		father = filepath.Dir(father)
	}
	err = dm.watcher.Add(father)
	if err != nil {
		log.Error("monitor father path failed.", zap.String("path", path), zap.Error(err),
			zap.String("father", father))
		return err
	}
	log.Info("add dir monitor.", zap.String("father", father), zap.String("file", file))
	dm.files.Store(filepath.Clean(file), true)

	dm.fileLock.Lock()
	dm.fileDirs[filepath.Clean(path)] = filepath.Clean(file)
	dm.fileLock.Unlock()

	return
}

// 监控路径
func (dm *DiskMonitor) monitorPath(path string, init bool) error {
	log := dm.opts.Logger.New("DiskMonitor.monitorPath")
	// 扫描目录下所有符合的文件
	allFiles, err := getAllFileWithExt(path, dm.opts.FileExt)
	if err != nil {
		log.Error("scan dir files error.", zap.Error(err), zap.String("path", path), zap.String("ext", dm.opts.FileExt))
		return err
	}
	// 加载所有文件
	for _, file := range allFiles {
		err = dm.loadFileAndNotify(file, "")
		if err != nil {
			log.Error("load dir files error.", zap.Error(err), zap.String("path", path), zap.String("ext", dm.opts.FileExt),
				zap.String("file", file), zap.Strings("allFiles", allFiles))
			if init {
				return err
			}
		}
	}
	// 是否已经监控
	if dm.isMonitorPath(path) {
		return nil
	}

	// 监控父目录 - 防止监控的目录本身进行修改(本身是链接,或者删除目录后重新添加的情况)
	father := filepath.Dir(path)
	// 递归获取连接目录的上一级目录
	var symlik string
	for {
		symlik, err = filepath.EvalSymlinks(father)
		if err != nil {
			log.Error("get real path failed.", zap.String("src", path), zap.String("father", father), zap.Error(err))
		}
		if symlik == father {
			break
		}
		father = filepath.Dir(father)
	}
	err = dm.watcher.Add(father)
	if err != nil {
		log.Error("monitor father path failed.", zap.String("path", path), zap.Error(err),
			zap.String("father", father))
		return err
	}
	log.Info("add dir monitor.", zap.String("father", father), zap.String("path", path))
	// 保存监控目录过滤
	dm.paths.Store(path, true)
	return nil
}

func (dm *DiskMonitor) notifyDirChange(path string) (err error) {
	log := dm.opts.Logger.New("DiskMonitor.notifyDirChange")
	dm.opts.Notifier.Prepare(dm.opts.Logger, path)
	//
	dm.fileLock.Lock()
	for dir, file := range dm.fileDirs {
		if hasPrefix(dir, path) {
			err = dm.loadFileAndNotify(file, "")
			if err != nil {
				log.Error("notify file change failed.", zap.String("path", path), zap.String("file", file))
			}
		}
	}
	dm.fileLock.Unlock()
	dm.opts.Notifier.Finish(dm.opts.Logger, path)
	return nil
}

func (dm *DiskMonitor) monitorChange(ctx context.Context, ch <-chan struct{}) {
	log := dm.opts.Logger.New("DiskMonitor.monitorChange")
	watcher := dm.watcher
	var list []*retryFile
	ticker := time.NewTicker(dm.opts.RetryInterval)
	defer ticker.Stop()

	var err error
	var absPath string
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Debug("recv not ok", zap.Any("event", event), zap.Bool("ok", ok))
				return
			}
			log.Debug("recv event", zap.Any("event", event))
			// maybe remove event. ignore.
			if !fileExists(event.Name) || event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Debug("ignore event.not exists or remove.", zap.Any("event", event))
				break
			}
			// 转换绝对路径
			absPath, err = filepath.Abs(event.Name)
			if err != nil {
				log.Error("convert abs failed.", zap.Error(err), zap.String("src", event.Name), zap.String("abs", absPath))
				break
			}
			// 是否目录
			if fileIsDir(absPath) {
				// 如果是监控的目录.那么就通知
				if dm.isMonitorPath(absPath) {
					dm.opts.Notifier.Prepare(dm.opts.Logger, absPath)
					dm.monitorPath(absPath, false)
					dm.opts.Notifier.Finish(dm.opts.Logger, absPath)
				} else {
					dm.notifyDirChange(absPath)
				}
				break
			}
			if dm.isMonitorFile(absPath) {
				dm.opts.Notifier.Prepare(dm.opts.Logger, absPath)
				dm.loadFileAndNotify(absPath, "")
				dm.opts.Notifier.Finish(dm.opts.Logger, absPath)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Error("monitor end with error.", zap.Error(err))
				return
			}
			log.Error("monitor error.", zap.Error(err))

		case item, ok := <-dm.retry:
			if !ok {
				return
			}
			list = append(list, item)
		case <-ticker.C:
			newList := []*retryFile{}
			for _, item := range list {
				err = dm.loadFileAndNotify(item.file, item.md5sum)
				if err != nil {
					item.times++
					log.Error("retry file failed.", zap.String("file", item.file),
						zap.String("md5", item.md5sum), zap.Int("times", item.times))
					if item.times < dm.opts.RetryLimit {
						newList = append(newList, item)
					}
				}
			}
			list = newList
		case <-ch:
			return
		case <-ctx.Done():
			return
		}
	}
}

// getAllFileWithExt 获取某个目录下所有ext后缀的文件
func getAllFileWithExt(path, ext string) (files []string, err error) {
	return getAllFile(path, func(file string) bool {
		if ext == "*" {
			return false
		}
		return filepath.Ext(file) != ext
	})
}

// getAllFile 获取某个目录下所有文件
func getAllFile(path string, filter func(string) bool) (files []string, err error) {
	if !fileIsDir(path) {
		return
	}
	rd, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fmt.Println("scan dir:", path+"/"+fi.Name())
			// fmt.Printf("[%s]\n", path+"/"+fi.Name())
			nfs, err := getAllFile(path+"/"+fi.Name()+"/", filter)
			if err != nil {
				return nil, err
			}
			files = append(files, nfs...)
		} else {
			if filter(fi.Name()) {
				fmt.Println("ignore file.", path+fi.Name())
				continue
			}
			// fmt.Println("scan file:", path+fi.Name())
			files = append(files, filepath.Clean(path+"/"+fi.Name()))
		}
	}
	return
}

// fileExists checks whether given <path> exist.
func fileExists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

// fileIsDir checks whether given <path> a directory.
func fileIsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func hasPrefix(str, pre string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(pre))
}
