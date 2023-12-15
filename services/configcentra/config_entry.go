package configcentra

import (
	"time"

	"log"

	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/app/bootstrap"
)

type ConfigCentra interface {
	// backend service interface
	Init(s app.Stoper) (err error)
	Start(s app.Stoper) error
	Stop()
	Finish()
	// value interface
	SetDefault(key string, doc string, value interface{})
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint16(key string) uint16
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
}

// 配置中心后端实现接口
var ConfigCentraBackend ConfigCentra

// ConfigValue 配置项 配置值
type ConfigValue interface {
	SetDefaultValue(vcfg ConfigCentra)
	RefreshValue(vcfg ConfigCentra)
}

type ConfigUpdateNotify func(ConfigCentra)
type FlagNotify func() error

// 配置中心
var gConfigManager = NewConfigService()

func init() {
	// config centra must start first of all.
	bootstrap.RegisterServiceByPriority(-1, gConfigManager) // config manager (load config from file)
}

// WatchConfigUpdate 监控配置更新
func WatchConfigUpdate(ntf ConfigUpdateNotify) {
	if gConfigManager.start.Load() {
		log.Panic("service already start, CAN NOT watch config any more.")
	}
	gConfigManager.updates = append(gConfigManager.updates, ntf)
}

// RegisterFlagNotify 注册flag处理
func RegisterFlagNotify(ntf FlagNotify) {
	if gConfigManager.start.Load() {
		log.Panic("service already start, CAN NOT register flag notify any more.")
	}
	gConfigManager.flags = append(gConfigManager.flags, ntf)
}

// RegisterConfig 注册配置
func RegisterConfig(cfg ConfigValue) {
	if gConfigManager.start.Load() {
		log.Panic("service already start, CAN NOT register config any more.")
	}
	gConfigManager.values = append(gConfigManager.values, cfg)
}
