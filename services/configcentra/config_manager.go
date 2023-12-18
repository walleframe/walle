package configcentra

import (
	"errors"
	"sync/atomic"

	"github.com/spf13/pflag"
	"github.com/walleframe/walle/app"
)

type cacheValue struct {
	cv  ConfigValue
	ntf []ConfigUpdateNotify
}

// ConfigCentraService 应用程序配置
type ConfigCentraService struct {
	app.NoopService
	start   atomic.Bool
	values  []cacheValue
	updates []ConfigUpdateNotify
	flags   []FlagNotify
}

func NewConfigService() *ConfigCentraService {
	return &ConfigCentraService{}
}

func (svc *ConfigCentraService) Name() string {
	return "config-manager"
}

func (svc *ConfigCentraService) Init(s app.Stoper) (err error) {
	svc.start.Store(true)
	// 命令行参数解析
	if !pflag.Parsed() {
		pflag.Parse()
	}
	// 命令行参数处理
	for _, ntf := range svc.flags {
		err = ntf()
		if err != nil {
			return err
		}
	}

	// must set config centra backend
	if ConfigCentraBackend == nil {
		return errors.New("not set config centra backend")
	}

	for _, v := range svc.values {
		ConfigCentraBackend.RegisterConfig(v.cv, v.ntf)
	}
	ConfigCentraBackend.WatchConfigUpdate(svc.updates)

	// config centra backend init
	err = ConfigCentraBackend.Init(s)
	if err != nil {
		return err
	}

	return
}

func (svc *ConfigCentraService) Start(s app.Stoper) error {
	return ConfigCentraBackend.Start(s)
}

func (svc *ConfigCentraService) Stop() {
	ConfigCentraBackend.Stop()
}

func (svc *ConfigCentraService) Finish() {
	ConfigCentraBackend.Finish()
}
