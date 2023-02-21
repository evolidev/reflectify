package reflectify

import (
	"github.com/mitchellh/mapstructure"
	"reflect"
	"runtime"
	"strings"
)

type ParamResolver func(rec *Reflection, parameter any) (any, bool)

func Reflect(v any) *Reflection {
	var val reflect.Value
	var t reflect.Type

	if _, ok := v.(reflect.Value); ok {
		val = v.(reflect.Value)
		t = val.Type()
	} else {
		val = reflect.ValueOf(v)
		t = reflect.TypeOf(v)
	}

	return &Reflection{
		v:               val,
		t:               t,
		resolvers:       make([]ParamResolver, 0),
		isReceiver:      false,
		element:         v,
		defaultResolver: defaultResolver,
	}
}

type Reflection struct {
	v               reflect.Value
	t               reflect.Type
	resolvers       []ParamResolver
	isReceiver      bool
	element         any
	defaultResolver ParamResolver
}

func (r *Reflection) Name() string {
	if r.t.Kind() == reflect.Func {
		funcName := r.functionName()
		paths := strings.Split(funcName, ".")

		return paths[len(paths)-1]
	}

	return r.reflectElem().Name()
}

func (r *Reflection) functionName() string {
	return runtime.FuncForPC(r.v.Pointer()).Name()
}

func (r *Reflection) reflectElem() reflect.Type {
	if r.t.Kind() == reflect.Ptr {
		return r.t.Elem()
	}

	return r.t
}

func (r *Reflection) Call(parameters ...interface{}) []reflect.Value {
	r.addFallbackResolver()

	callParams := r.buildInputParameters(parameters)

	return r.v.Call(callParams)
}

func (r *Reflection) CallMethod(s string, parameters ...interface{}) []reflect.Value {
	if r.t.Kind() == reflect.Func {
		return r.Call(parameters)
	}

	m := r.v.MethodByName(s)
	refl := Reflect(m)

	return refl.Call(parameters)
}

func (r *Reflection) HasReceiver() bool {
	if r.t.NumIn() == 0 {
		return false
	}

	first := r.t.In(0)
	if first.Kind() != reflect.Struct && first.Kind() != reflect.Ptr {
		return false
	}

	if _, ok := first.MethodByName(r.Name()); !ok {
		return false
	}

	return true
}

func (r *Reflection) AddResolver(resolver ParamResolver) {
	r.resolvers = append(r.resolvers, resolver)
}

func (r *Reflection) resolve(currentInputParam *Reflection, parameters []interface{}) reflect.Value {
	var param interface{}
	if len(parameters) > 0 {
		param = parameters[0]
	} else {
		param = nil
	}

	for _, resolver := range r.resolvers {
		resolved, paramUsed := resolver(currentInputParam, param)

		if resolved != nil {
			if paramUsed {
				if len(parameters) > 0 {
					parameters = parameters[1:]
				}
			}

			return reflect.ValueOf(resolved)
		}
	}

	return reflect.ValueOf(param)
}

func (r *Reflection) makeNewValueForInput(paramType reflect.Type) interface{} {
	if paramType.Kind() == reflect.Struct || paramType.Kind() == reflect.Ptr {
		destination := reflect.New(paramType).Elem().Interface()
		reflectValue := reflect.ValueOf(destination)
		t := reflectValue.Type()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		tmp := reflect.New(t)
		if paramType.Kind() != reflect.Ptr {
			tmp = tmp.Elem()
		}

		destination = tmp.Interface()

		return destination
	}

	return reflect.New(paramType).Elem().Interface()
}

func (r *Reflection) addFallbackResolver() {
	r.AddResolver(r.defaultResolver)
}

func (r *Reflection) buildInputParameters(parameters []interface{}) []reflect.Value {
	params := make([]reflect.Value, 0)
	cnt := len(params)

	for cnt < r.t.NumIn() {
		currentInputParam := Reflect(r.makeNewValueForInput(r.t.In(cnt)))
		if cnt == 0 && r.HasReceiver() {
			currentInputParam.isReceiver = true
		}

		cnt++

		resolved := r.resolve(currentInputParam, parameters)
		params = append(params, resolved)
	}

	return params
}

func (r *Reflection) InstanceOf(instance interface{}) bool {
	refl := Reflect(instance)

	match := refl.Name() == r.Name()

	if match {
		if r.t.Kind() == reflect.Ptr {
			match = refl.t.Kind() == reflect.Ptr
		} else {
			match = refl.t.Kind() != reflect.Ptr
		}
	}

	return match
}

func (r *Reflection) IsReceiver() bool {
	return r.isReceiver
}

func (r *Reflection) New() interface{} {
	r.element = r.makeNewValueForInput(r.t)

	return r.Element()
}

func (r *Reflection) Fill(input interface{}) interface{} {
	elem := r.element

	if !r.IsPointer() {
		mapstructure.Decode(input, &elem)
	} else {
		mapstructure.Decode(input, elem)
	}

	return elem
}

func (r *Reflection) IsPointer() bool {
	return r.t.Kind() == reflect.Ptr
}

func (r *Reflection) Element() interface{} {
	return r.element
}

var defaultResolver = func(rec *Reflection, parameter any) (any, bool) {
	tmp := rec.New()
	if parameter != nil {
		m := NewMapper(parameter)
		if rec.t.Kind() == reflect.Int {
			tmp = m.Int()
		} else if rec.t.Kind() == reflect.String {
			tmp = m.String()
		} else if rec.t.Kind() == reflect.Bool {
			tmp = m.Bool()
		} else {
			tmp = parameter
		}
	}

	return tmp, true
}
