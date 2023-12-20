package configcentra

import (
	"time"
)

type configItem struct {
	set func(vcfg ConfigCentra)
	get func(vcfg ConfigCentra) error
}

func (cfg *configItem) SetDefaultValue(vcfg ConfigCentra) {
	cfg.set(vcfg)
}
func (cfg *configItem) RefreshValue(vcfg ConfigCentra) error {
	return cfg.get(vcfg)
}

// String register its value ,type and desc to config centra,and auto update when it values is changed by remote.
func String(ptr *string, key string, def string, doc string, updateNtfs ...func(val string)) {
	*ptr = def
	RegisterConfig(&configItem{
		set: func(vcfg ConfigCentra) {
			vcfg.SetDefault(key, doc, def)
		},
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetString(key)
			if err != nil {
				return err
			}

			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetBool(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetInt(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetInt32(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetInt64(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetUint(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetUint32(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetUint64(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetFloat64(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetTime(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetDuration(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetIntSlice(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
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
		get: func(vcfg ConfigCentra) error {
			val, err := vcfg.GetStringSlice(key)
			if err != nil {
				return err
			}
			*ptr = val
			for _, ntf := range updateNtfs {
				ntf(val)
			}
			return nil
		},
	})
}
