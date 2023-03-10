package reflectify

import (
	"errors"
	"strings"
	"testing"
)

func TestReflectOfReflection(t *testing.T) {
	refl := Reflect(func(test string) {})
	refl2 := Reflect(refl)

	if refl2.IsStruct() {
		t.Errorf("reflected element should still be a func")
	}
	if len(refl2.Params()) != 1 {
		t.Errorf("reflection should be the same")
	}
}

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

func TestMethods(t *testing.T) {
	t.Run("Methods should return all struct methods", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		methods := refl.Methods()

		if len(methods) != 3 {
			t.Errorf("Current len of %d does not match expected %d", len(methods), 3)
		}
	})

	t.Run("Result of methods should be able to return input params", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		methods := refl.Methods()

		method := methods["TestWithScalarParam"]
		params := method.Params()

		if len(params) != 1 {
			t.Errorf("given count %d of params are not matching expected %d", len(params), 1)
		}

		if !params[0].IsScalar() {
			t.Errorf("wrong type assertion")
		}
	})

	t.Run("Calling params on struct reflection should return empty result", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		params := refl.Params()

		if len(params) > 0 {
			t.Errorf("params should be empty")
		}
	})

	t.Run("Calling methods on a function should return it self", func(t *testing.T) {
		refl := Reflect(func() {})

		methods := refl.Methods()

		if len(methods) != 1 {
			t.Errorf("current len of %d does not match with expected %d", len(methods), 1)
		}
	})
}

func TestMethodByName(t *testing.T) {
	t.Run("method by name should return the method of the struct", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		method := refl.MethodByName("TestWithScalarParam")

		if method == nil {
			t.Errorf("Method should exists")
		}
	})

	t.Run("method by name should return the method if reflected element is a func", func(t *testing.T) {
		refl := Reflect(func() string { return "test" })

		method := refl.MethodByName("TestWithScalarParam")
		if method == nil {
			t.Errorf("func should exists either way")
		}

		result := method.Call()

		if result[0].String() != "test" {
			t.Errorf("method could not be called")
		}
	})

	t.Run("method by name should return nil if method does not exists", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		method := refl.MethodByName("DoesNotExists")

		if method != nil {
			t.Errorf("Method should be nil")
		}
	})
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

	t.Run("if resolver returns an error then it should be returned", func(t *testing.T) {
		refl := Reflect(func(testStruct *TestStruct, tmp int) int { return tmp })
		refl.AddResolver(func(definition *Reflection, param any) (any, bool) {
			return errors.New("failed"), true
		})

		result := refl.Call(5)

		if len(result) != 2 {
			t.Errorf("expected results should be of length %d. Current length: %d", 2, len(result))
		}

		if _, ok := result[1].Interface().(error); !ok {
			t.Errorf("expected error as 2nd result. %v given", result[1].Interface())
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

func TestFullName(t *testing.T) {
	t.Run("fullname should prepend package path", func(t *testing.T) {
		refl := Reflect(TestStruct{})

		fullName := refl.FullName()

		expected := "github.com/evolidev/reflectify/TestStruct"
		if fullName != expected {
			t.Errorf("name expected to be '%s'. '%s' given", expected, fullName)
		}
	})

	t.Run("fullname should prepend package path even for pointer", func(t *testing.T) {
		refl := Reflect(&TestStruct{})

		fullName := refl.FullName()

		expected := "github.com/evolidev/reflectify/TestStruct"
		if fullName != expected {
			t.Errorf("name expected to be '%s'. '%s' given", expected, fullName)
		}
	})

	t.Run("fullname should prepend package path even for func", func(t *testing.T) {
		refl := Reflect(TestStruct.TestWithMultiParam)

		fullName := refl.FullName()

		expected := "github.com/evolidev/reflectify/TestStruct:TestWithMultiParam"
		if fullName != expected {
			t.Errorf("name expected to be '%s'. '%s' given", expected, fullName)
		}
	})

	t.Run("fullname should prepend package path and even for anonymous functions", func(t *testing.T) {
		refl := Reflect(func() {})

		fullName := refl.FullName()

		expected := "github.com/evolidev/reflectify.TestFullName.func"
		if !strings.Contains(fullName, expected) {
			t.Errorf("name expected to be '%s'. '%s' given", expected, fullName)
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

	t.Run("if parameter is a struct than it should be passed to through", func(t *testing.T) {
		refl := Reflect(func(testStruct *TestStruct) string {
			return testStruct.Field1
		})

		result := refl.CallMethod("Handle", &TestStruct{Field1: "test"})

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

func (s TestStruct) TestWithScalarParam(t string) string {
	return "hello"
}

func (s TestStruct) TestWithMultiParam(ts *TestStruct2, t string) string {
	return "hello"
}

func (s TestStruct) private() string {
	return "hello"
}

type TestStruct2 struct {
}

func testFunc(param1 string, param2 TestStruct2) string {
	return "test"
}
