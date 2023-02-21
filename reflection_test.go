package reflectify

import (
	"testing"
)

func TestInstanceOf(t *testing.T) {
	t.Run("instance of should return true if current reflected struct is of the same type", func(t *testing.T) {
		tmp := Reflect(TestStruct{})

		result := tmp.InstanceOf(TestStruct{})

		if !result {
			t.Errorf("Instance check is wrong")
		}
	})

	t.Run("instance of should return false if current reflect type is pointer and check value is not", func(t *testing.T) {
		tmp2 := Reflect(&TestStruct{})

		result := tmp2.InstanceOf(TestStruct{})

		if result {
			t.Errorf("Instance check is wrong")
		}
	})
}

func TestIsStruct(t *testing.T) {
	refl := Reflect(TestStruct{})

	result := refl.IsStruct()

	if !result {
		t.Errorf("Is struct should return true if reflected element is a struct")
	}
}

func TestIsScalar(t *testing.T) {
	refl := Reflect("")

	result := refl.IsScalar()

	if !result {
		t.Errorf("Is scalar should return true if reflected element is a scalar type")
	}
}

func TestFill(t *testing.T) {
	t.Run("fill should fill none pointer struct", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		test := make(map[string]interface{})
		test["field1"] = "test"

		result := refl.Fill(test)

		if result.(TestStruct).Field1 != "test" {
			t.Errorf("damn")
		}
	})

	t.Run("fill should fill pointer struct", func(t *testing.T) {
		refl := Reflect(&TestStruct{})

		test := make(map[string]interface{})
		test["field1"] = "test"

		result := refl.Fill(test)

		if result.(*TestStruct).Field1 != "test" {
			t.Errorf("damn")
		}
	})
}

func TestCall(t *testing.T) {
	t.Run("call should call the function", func(t *testing.T) {
		called := false
		refl := Reflect(func() { called = true })

		refl.Call()

		if !called {
			t.Errorf("function got not called")
		}
	})

	t.Run("call should inject given int parameter", func(t *testing.T) {
		tmp := 0
		refl := Reflect(func(param int) { tmp = param })

		refl.Call(1)

		if tmp != 1 {
			t.Errorf("parameter was not given")
		}
	})

	t.Run("call should inject given string parameter", func(t *testing.T) {
		tmp := ""
		refl := Reflect(func(param string) { tmp = param })

		refl.Call("test")

		if tmp != "test" {
			t.Errorf("parameter was not given")
		}
	})

	t.Run("call should inject given bool parameter", func(t *testing.T) {
		tmp := false
		refl := Reflect(func(param bool) { tmp = param })

		refl.Call(true)

		if tmp != true {
			t.Errorf("parameter was not given")
		}
	})

	t.Run("call should call method", func(t *testing.T) {
		refl := Reflect(TestStruct.Test)

		result := refl.Call()

		if result[0].String() != "hello" {
			t.Errorf("method does not get called")
		}
	})

	t.Run("call should call resolver who can handle current param", func(t *testing.T) {
		refl := Reflect(func(testStruct *TestStruct) int { return testStruct.field2 })
		refl.AddResolver(func(definition *Reflection, param any) (any, bool) {
			if definition.InstanceOf(&TestStruct{}) {
				ts := definition.New().(*TestStruct)
				ts.field2 = param.(int)

				return ts, true
			}

			return nil, false
		})

		result := refl.Call(5)

		if result[0].Int() != 5 {
			t.Errorf("wrong value given")
		}
	})

	t.Run("call should resolve param and should inject other params if not custom resolved", func(t *testing.T) {
		refl := Reflect(func(testStruct *TestStruct, tmp int) int { return tmp })
		refl.AddResolver(func(definition *Reflection, param any) (any, bool) {
			if definition.InstanceOf(&TestStruct{}) {
				return definition.New(), false
			}

			return nil, false
		})

		result := refl.Call(5)

		if result[0].Int() != 5 {
			t.Errorf("current value '%d' given. expected: %d", result[0].Int(), 5)
		}
	})

	t.Run("call should map value to desired one", func(t *testing.T) {
		tmp := 0
		refl := Reflect(func(param int) { tmp = param })

		refl.Call("1")

		if tmp != 1 {
			t.Errorf("parameter was not given")
		}
	})

	t.Run("call should map value to desired one", func(t *testing.T) {
		tmp := 0
		s := struct {
		}{}
		refl := Reflect(func(param struct{}) { tmp = 1 })

		refl.Call(s)

		if tmp != 1 {
			t.Errorf("parameter was not given")
		}
	})

	t.Run("call with nil", func(t *testing.T) {
		tmp := 0
		refl := Reflect(func(param []string) { tmp = 1 })

		refl.Call(nil)

		if tmp != 1 {
			t.Errorf("parameter was not given")
		}
	})

	t.Run("call with no default resolver should try to call with value of param", func(t *testing.T) {
		called := false
		refl := Reflect(func(test string) { called = true })
		refl.defaultResolver = func(rec *Reflection, parameter any) (any, bool) {
			return nil, false
		}

		refl.Call("test")

		if !called {
			t.Errorf("func got not called")
		}
	})
}

func TestIsReceiver(t *testing.T) {
	t.Run("is receiver should return false if it is a struct", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		result := refl.IsReceiver()

		if result {
			t.Errorf("a simple struct should not be marked as receiver")
		}
	})

	t.Run("is receiver should return true if definition in resolver is the receiver", func(t *testing.T) {
		isReceiver := false
		refl := Reflect(TestStruct.Test)
		refl.AddResolver(func(rec *Reflection, parameter any) (any, bool) {
			if rec.IsReceiver() {
				isReceiver = true

				return rec.New(), false
			}

			return nil, false
		})

		refl.Call()

		if !isReceiver {
			t.Errorf("receiver should be marked as receiver")
		}
	})
}

func TestCallMethod(t *testing.T) {
	t.Run("method of struct should be called", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		result := refl.CallMethod("Test")

		if result[0].String() != "hello" {
			t.Errorf("func not called")
		}
	})

	t.Run("if method is function then it should be simple called", func(t *testing.T) {
		refl := Reflect(func() string { return "test" })

		result := refl.CallMethod("")

		if result[0].String() != "test" {
			t.Errorf("func not called")
		}
	})
}

func TestReceiver(t *testing.T) {
	t.Run("Has receiver should return false if func does not have a receiver", func(t *testing.T) {
		refl := Reflect(func() {})

		if refl.HasReceiver() {
			t.Errorf("wrong assertions of receiver existence")
		}
	})

	t.Run("Has receiver should return true if func does have a receiver", func(t *testing.T) {
		refl := Reflect(TestStruct.Test)

		if !refl.HasReceiver() {
			t.Errorf("wrong assertions of receiver existence")
		}
	})

	t.Run("Has receiver should return false if func does not have a receiver but input parameters", func(t *testing.T) {
		refl := Reflect(func(test string) {})

		if refl.HasReceiver() {
			t.Errorf("wrong assertions of receiver existence")
		}
	})

	t.Run("Has receiver should return false if func does not have a receiver but a struct parameter as first param", func(t *testing.T) {
		refl := Reflect(func(test TestStruct) {})

		if refl.HasReceiver() {
			t.Errorf("wrong assertions of receiver existence")
		}
	})
}

func TestElement(t *testing.T) {
	ts := TestStruct{Field1: "hello"}
	refl := Reflect(ts)

	elem := refl.Element()

	if elem.(TestStruct).Field1 != "hello" {
		t.Errorf("Field has value '%s' but expected '%s'", elem.(TestStruct).Field1, "hello")
	}
}

func TestNew(t *testing.T) {
	t.Run("new should return a new instance of given reflected item", func(t *testing.T) {
		ts := TestStruct{Field1: "hello"}
		refl := Reflect(ts)

		elem := refl.New()

		if elem.(TestStruct).Field1 != "" {
			t.Errorf("Field has value '%s' but expected empty", elem.(TestStruct).Field1)
		}
	})

	t.Run("new should reset element to a new instance", func(t *testing.T) {
		ts := TestStruct{Field1: "hello"}
		refl := Reflect(ts)
		elem := refl.New()

		tmp := refl.Element()

		if elem.(TestStruct).Field1 != "" {
			t.Errorf("Field has value '%s' but expected empty", elem.(TestStruct).Field1)
		}
		if tmp.(TestStruct).Field1 != "" {
			t.Errorf("Field of element has value '%s' but expected empty", elem.(TestStruct).Field1)
		}
	})

}

func TestName(t *testing.T) {
	t.Run("name of struct should be returned", func(t *testing.T) {
		s := newTestStruct()
		reflection := Reflect(s)

		if reflection.Name() != "TestStruct" {
			t.Errorf("current name '%s' does not match expected '%s'", reflection.Name(), "TestStruct")
		}
	})

	t.Run("name of struct should be returned even if it is a pointer", func(t *testing.T) {
		s := newTestStructPointer()
		reflection := Reflect(s)

		if reflection.Name() != "TestStruct" {
			t.Errorf("current name '%s' does not match expected '%s'", reflection.Name(), "TestStruct")
		}
	})

	t.Run("name of func should be returned", func(t *testing.T) {
		s := testFunc
		reflection := Reflect(s)

		if reflection.Name() != "testFunc" {
			t.Errorf("current name '%s' does not match expected '%s'", reflection.Name(), "testFunc")
		}
	})
}

func newTestStruct() TestStruct {
	return TestStruct{
		Field1: "test",
		field2: 2,
		field3: func() {},
		field4: TestStruct2{},
	}
}

func newTestStructPointer() *TestStruct {
	s := newTestStruct()

	return &s
}

type TestStruct struct {
	Field1 string
	field2 int
	field3 func()
	field4 TestStruct2
}

func (s TestStruct) Test() string {
	return "hello"
}

type TestStruct2 struct {
}

func testFunc(param1 string, param2 TestStruct2) string {
	return "test"
}
