package livr

import "strings"

var defaultAutoTrim = false

type isAutoTrim uint8

func (at isAutoTrim) Bool() bool {
	if at == DoTrim {
		return true
	}
	return false
}

const (
	// Nil - is non ommited/disabled validator auto trim mode.
	Nil isAutoTrim = iota
	// NotTrim - is disabled validator auto trim mode.
	NotTrim
	// DoTrim - is enabled validator auto trim mode.
	DoTrim
)

// SetAutoTrim - turn on/off data auto trim.
func SetAutoTrim(at isAutoTrim) {
	defaultAutoTrim = at.Bool()
}

func autoTrim(data interface{}) interface{} {
	switch d := data.(type) {
	case string:
		return strings.TrimSpace(d)
	case map[string]interface{}:
		var trimmedData = make(map[string]interface{})
		for key, val := range d {
			trimmedData[key] = autoTrim(val)
		}
		return trimmedData
	case []interface{}:
		var trimmedData []interface{}
		for val := range d {
			trimmedData = append(trimmedData, autoTrim(val))
		}
		return trimmedData
	default:
		return d
	}
}
