package xlsxmgr

import (
	"crypto/md5"
	"fmt"

	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

// ConfigItem 配置项
type ConfigItem struct {
	basename  string              // 本地json文件名
	fromXlsx  string              // xlsx文件来源
	fromSheet string              // 原始sheet名称
	version   string              // 当前版本-md5
	parser    DataParser          // 数据解析
	loader    ConfigLoader        // 数据引用
	externs   []externLoaderCache // 附加引用
	mixConfig []mixLoaderCache    // 混合数据接口

	mgr *XlsxConfig
}

func (cfg *ConfigItem) GetFileName(ext string) string {
	return cfg.basename + ext
}

func (cfg *ConfigItem) GetBaseName() string {
	return cfg.basename
}

// LoadConfigFromData 加载字节流到导出配置结构体中
func (cfg *ConfigItem) LoadConfigFromData(data []byte) error {
	logger := cfg.logger().New("ConfigItem.loadConfigFromData")
	// 校验文件是否有改动
	newVersion := fmt.Sprintf("%x", md5.Sum(data))
	if newVersion == cfg.version {
		logger.Warn("file no change",
			zap.String("file", cfg.basename),
		)
		return nil
	}

	baseContainer := cfg.loader.NewContainer()

	var err error
	// 解析数据到容器中
	if cfg.parser != nil {
		// 自定义解析文件
		err = cfg.parser.UnmarshalXlsxData(data, baseContainer)
	} else {
		err = cfg.mgr.plugin.UnmarshalXlsxData(data, baseContainer)
	}
	if err != nil {
		logger.Error("unmarshal data failed",
			zap.Error(err),
			zap.String("file", cfg.basename),
		)
		return err
	}
	// 检查数据
	if err = cfg.loader.Check(baseContainer); err != nil {
		logger.Error("check data failed",
			zap.Error(err),
			zap.String("file", cfg.basename),
		)
		return fmt.Errorf("file[%s] on checked failed[%s]", cfg.basename, err)
	}

	// 追加数据解析
	cache := make([]innerLoaderCache, 0, len(cfg.externs))
	for _, ext := range cfg.externs {
		container, err := ext.loader.Parse(baseContainer)
		if err != nil {
			logger.Error("parse extern config failed",
				zap.String("file", cfg.basename),
				zap.String("tag", ext.tag),
				zap.Error(err),
			)
			return err
		}
		err = ext.loader.Check(container)
		if err != nil {
			logger.Error("check extern config failed",
				zap.String("file", cfg.basename),
				zap.String("tag", ext.tag),
				zap.Error(err),
			)
			return err
		}

		cache = append(cache, innerLoaderCache{
			swap: ext.loader.Swap,
			data: container,
		})
	}

	// 最终交换数据
	cfg.loader.Swap(baseContainer)
	for _, v := range cache {
		v.swap(v.data)
	}
	cfg.version = newVersion

	logger.Info("load data success",
		zap.String("file", cfg.basename),
	)
	return nil
}

// BuildMixData 构建混合数据
func (cfg *ConfigItem) BuildMixData() {
	logger := cfg.logger().New("ConfigItem.BuildMixData")
	for _, mix := range cfg.mixConfig {
		err := mix.buildMixData()
		if err != nil {
			logger.Error("parse extern config failed",
				zap.String("file", cfg.basename),
				zap.String("tag", mix.tag),
				zap.Error(err),
			)
		}
	}
}

func (cfg *ConfigItem) logger() *zaplog.Logger {
	return cfg.mgr.logger
}

type innerLoaderCache struct {
	swap func(new interface{})
	data interface{}
}

type mixLoaderCache struct {
	tag          string
	buildMixData func() error
}

type externLoaderCache struct {
	tag    string
	loader ConfigAppendLoader
}
