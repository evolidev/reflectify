# reflectify
Simple reflect helper

Reflectify makes it easy to call any sort of functions. 
Lets start with a simple example.
```go
refl := Reflect(func() string { return "string" })

result := refl.Call()

// prints "string"
fmt.Println(result[0].String())
```

If your function has parameter, don`t worry we will also handle that for you. 
Just pass the parameters to the `Call` method. 
```go
refl := Reflect(func(param string) string { return param })

result := refl.Call("hello")

// prints "hello"
fmt.Println(result[0].String())
```

Get rid of manually convert the parameters to their desired types.
```go
refl := Reflect(func(param int) int { return param })

result := refl.Call("1")

// prints 1
fmt.Println(result[0].Int())
```

Sometimes you have more complex parameter resolving logic. 
For example your function requires a DB model but you pass an int to call. 
For that kind of use cases you can provide custom resolver functions.
```go
refl := Reflect(func(model *MyModel) *MyModel { return param })
refl.AddResolver(func(rec *Reflection, parameter any) (any, bool) {
    if rec.InstanceOf(&MyModel{}) {
        model := rec.Elem()
        DB.GetById(model, parameter.(int))
		
        return model, true
    }   
	
    return nil, false
})

result := refl.Call(1)

// prints your model
fmt.Println(result[0].Interface().(*MyModel))
```
Lets discuss above code. 
First we create a new reflection struct of the function. 
Next we add a resolver. A resolver get as first parameter an instance of `*Reflection` struct. 
This holds the information about the functions parameter. In that case `*MyModel`. 
The second parameter is the next parameter given through the `Call` function. 
The first return type should be the resolved value for the function param. 
The second return type will indicate if the param was used or not. 
We will dig into it a bit later why this is important. 
In the body we use the helper method `InstanceOf` to check if the types matches. 
If so we fetch the model from the DB and return the value else we return nil and false. 


Now lets check an example when we should return false for the 2nd parameter in the resolver function. 
```go
refl := Reflect(func(aStruct *TestStruct, param string) *MyModel { return param })
refl.AddResolver(func(rec *Reflection, parameter any) (any, bool) {
    if rec.InstanceOf(&TestStruct{}) {
        return rec.New(), false
    }   
	
    return nil, false
})

result := refl.Call("test")

// prints your model
fmt.Println(result[0].String())
```
As you can see the `*TestStruct` should simple be injected by whatever logic you wish. 
But it is not coupled with the params you provided. 

Often you will generally resolve something only for the `receiver of the function`. 
```go
type TestStruct struct {
	Test string
}

func(t TestStruct) MyFunc() string {
	return t.Test
}

refl := Reflect(TestStruct.MyFunc)
refl.AddResolver(func(rec *Reflection, parameter any) (any, bool) {
    if rec.IsReceiver() {
		myStuff := make(map[string]interface{})
		myStuff["test"] = "hello"
		
        return rec.Fill(myStuff), false
    }   
	
    return nil, false
})

result := refl.Call("test")

// prints your model
fmt.Println(result[0].String())
```

`Fill` will fill the current element. 
With `New` you get new instance of the current reflected item.
`New` also resets the element which you will get with `Eleement`. 

```go
refl := Reflect(TestStruct{Field1: "hello"})
// prints "hello"
fmt.Println(refl.Element().(TestStruct).Field1)

newElement := refl.New()
// prints ""
fmt.Println(newElement.(TestStruct).Field1)

element := refl.Element()
// prints "" as the internal element got reset
fmt.Println(newElement.(TestStruct).Field1)
```