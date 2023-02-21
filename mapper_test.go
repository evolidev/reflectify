package reflectify

import "testing"

func TestMapToString(t *testing.T) {
	t.Run("map int to string", func(t *testing.T) {
		mapper := NewMapper(2)

		result := mapper.String()

		if result != "2" {
			t.Errorf("could not map to string")
		}
	})

	t.Run("map bool to string", func(t *testing.T) {
		mapper := NewMapper(true)

		result := mapper.String()

		if result != "true" {
			t.Errorf("could not map bool '%v' to string '%s'. Got: %v", true, "true", result)
		}
	})

	t.Run("map string to string", func(t *testing.T) {
		mapper := NewMapper("test")

		result := mapper.String()

		if result != "test" {
			t.Errorf("could not map string '%v' to string '%s'. Got: %v", "test", "test", result)
		}
	})
}

func TestMapToInt(t *testing.T) {
	t.Run("map int to int", func(t *testing.T) {
		mapper := NewMapper(2)

		result := mapper.Int()

		if result != 2 {
			t.Errorf("could not map to int")
		}
	})

	t.Run("map bool to int", func(t *testing.T) {
		mapper := NewMapper(true)

		result := mapper.Int()

		if result != 1 {
			t.Errorf("could not map bool '%v' to int '%d'. Got: %v", true, 1, result)
		}
	})

	t.Run("map bool to int", func(t *testing.T) {
		mapper := NewMapper(false)

		result := mapper.Int()

		if result != 0 {
			t.Errorf("could not map bool '%v' to int '%d'. Got: %v", false, 0, result)
		}
	})

	t.Run("map string to int", func(t *testing.T) {
		mapper := NewMapper("1")

		result := mapper.Int()

		if result != 1 {
			t.Errorf("could not map string '%v' to int '%d'. Got: %v", "1", 1, result)
		}
	})

	t.Run("map string to int", func(t *testing.T) {
		mapper := NewMapper("test")

		result := mapper.Int()

		if result != 0 {
			t.Errorf("could not map string '%v' to int '%d'. Got: %v", "1", 1, result)
		}
	})
}

func TestMapToBool(t *testing.T) {
	t.Run("map int to bool", func(t *testing.T) {
		mapper := NewMapper(2)

		result := mapper.Bool()

		if !result {
			t.Errorf("could not map to bool")
		}
	})

	t.Run("map bool to int", func(t *testing.T) {
		mapper := NewMapper(true)

		result := mapper.Bool()

		if !result {
			t.Errorf("could not map bool '%v' to bool '%v'. Got: %v", true, true, result)
		}
	})

	t.Run("map string to int", func(t *testing.T) {
		mapper := NewMapper("1")

		result := mapper.Bool()

		if !result {
			t.Errorf("could not map string '%v' to bool '%v'. Got: %v", "1", true, result)
		}
	})

	t.Run("map string to int", func(t *testing.T) {
		mapper := NewMapper(struct{}{})

		result := mapper.Bool()

		if result {
			t.Errorf("could not map struct '%v' to bool '%v'. Got: %v", struct{}{}, false, result)
		}
	})
}
