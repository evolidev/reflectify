package reflectify

import (
	"strconv"
)

func NewMapper(value any) *Mapper {
	return &Mapper{value: value}
}

type Mapper struct {
	value any
}

func (m *Mapper) String() string {
	switch m.value.(type) {
	case int:
		return strconv.Itoa(m.value.(int))
	case bool:
		return strconv.FormatBool(m.value.(bool))
	default:
		return m.value.(string)
	}
}

func (m *Mapper) Int() int {
	switch m.value.(type) {
	case bool:
		tmp := m.value.(bool)
		if tmp {
			return 1
		}

		return 0
	case string:
		result, err := strconv.Atoi(m.value.(string))
		if err != nil {
			return 0
		}

		return result
	default:
		return m.value.(int)
	}
}

func (m *Mapper) Bool() bool {
	switch m.value.(type) {
	case bool:
		return m.value.(bool)
	case int:
		return m.value.(int) > 0
	case string:
		return m.value.(string) != ""
	default:
		return false
	}
}
