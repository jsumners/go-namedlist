# namedlist

The purpose of this library is to provide methods for generating a slice
of `sql.NamedArg` from a `struct`. The bones are present, but I decided that
it isn't worth the effort at this time.

## Example

```go
type Person struct {
	Name string
	Email string
	Ignored string `db:"-"`
}

person := Person{
	Name: "Foo"
	Email: "foo@example.com"
	Ignored: "something"
}

namedList, _ := namedlist.New()
list, _ := namedList.FromStruct(person)

db.Query('insert statement', list...)
```

## Issues

+ Can't hoist nested fields to a top level name, e.g. `Parent.Child.Name`
can only be something like `child.name` when we probably want something like
`child1_name`
+ 
