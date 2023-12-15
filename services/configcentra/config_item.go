package configcentra

import (
	"time"
)

type configItem struct {
	set func(vcfg ConfigCentra)
	get func(vcfg ConfigCentra)
}

func (cfg *configItem) SetDefaultValue(vcfg ConfigCentra) {
	cfg.set(vcfg)
}
func (cfg *configItem) RefreshValue(vcfg ConfigCentra) {
	cfg.get(vcfg)
}

// String register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func String(ptr *string, key string, def string, doc string, updateNtfs ...func(val string)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetString(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Bool register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Bool(ptr *bool, key string, def bool, doc string, updateNtfs ...func(val bool)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetBool(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Int register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Int(ptr *int, key string, def int, doc string, updateNtfs ...func(val int)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetInt(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Int32 register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Int32(ptr *int32, key string, def int32, doc string, updateNtfs ...func(val int32)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetInt32(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Int64 register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Int64(ptr *int64, key string, def int64, doc string, updateNtfs ...func(val int64)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetInt64(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Uint register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Uint(ptr *uint, key string, def uint, doc string, updateNtfs ...func(val uint)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetUint(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Uint32 register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Uint32(ptr *uint32, key string, def uint32, doc string, updateNtfs ...func(val uint32)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetUint32(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Uint64 register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Uint64(ptr *uint64, key string, def uint64, doc string, updateNtfs ...func(val uint64)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetUint64(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Float64 register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Float64(ptr *float64, key string, def float64, doc string, updateNtfs ...func(val float64)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetFloat64(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Time register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Time(ptr *time.Time, key string, def time.Time, doc string, updateNtfs ...func(val time.Time)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetTime(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// Duration register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func Duration(ptr *time.Duration, key string, def time.Duration, doc string, updateNtfs ...func(val time.Duration)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetDuration(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// IntSlice register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func IntSlice(ptr *[]int, key string, def []int, doc string, updateNtfs ...func(val []int)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetIntSlice(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}

// StringSlice register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func StringSlice(ptr *[]string, key string, def []string, doc string, updateNtfs ...func(val []string)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) {
			val := vcfg.GetStringSlice(key)
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
		},
	})
}
