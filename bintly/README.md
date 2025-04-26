# Bintly (super fast binary serialization for go) 

[![GoReportCard](https://goreportcard.com/badge/github.com/viant/bintly)](https://goreportcard.com/report/github.com/viant/bintly)
[![GoDoc](https://godoc.org/github.com/viant/bintly?status.svg)](https://godoc.org/github.com/viant/bintly)

This library is compatible with Go 1.11+

Please refer to [`CHANGELOG.md`](CHANGELOG.md) if you encounter breaking changes.

- [Motivation](#motivation)
- [Introduction](#introduction)
- [Usage](#usage)
- [Custom encoding/decoding](#custom-encodingdecoding)
- [Bugs](#bugs)
- [Contribution](#contributing-to-bintly)
- [License](#license)

## Motivation

The goal of library to provide super fast binary oriented decoding and encoding capability for any go data type, critical
for low latency applications.


## Introduction

Typical streamlined binary serialization format store primitive types with their native size, and all collection type
got pre seeded with the repeated data size. Imagine the follow struct:

```go
type Employee struct {
	ID int
	Name string
	RolesIDs []int
	Titles []string
    DeptIDs []int
}

var emp := Employee{
    ID: 100,
    Name: "test",
    RolesIDs: []int{1000,1002,1003},
    Titles: []string{"Lead", "Principal"},
    DeptIDs: []int{10,13},
}
```
This maps to the following binary stream representation:
```
100,4,test,3,1000,1002,1003,2,4,Lead,9,Principal,2,10,13
```

In examples presented coma got preserved only for visualisation, also numeric/alphanumerics usage is for simplification.

When decoding this binary format each repeated type requires new memory allocation, in this case 6 allocations:
3 for slices, and 3 for string type. 

Since it's possible to copy any primitive slice to memory back and forth, we can go about binary serialization way faster than the originally presented approach.
Instead of allocation memory for each repeated type (string,slice), we could simply reduce number allocation to number of 
primitive data type used + 1 to track allocations.
In that case binary data stream for emp variable will look like the following. 

```yaml
alloc: [4,3,2,4,9,2] 
ints: [100,1000,1002,1003,10,13]
uint8s: [test,Lead,Principal]
```

The performance of this library comes by reducing memory allocation. 
In addition, all native data got copied to bytes back and forth without any conversion, 
improving additionally speed. Since there is no specific data conversion used,
all serialized data uses native golang numeric data type representation. 


## Usage

```go
func Example_Marshal() {
	emp := Employee{
		ID:       100,
		Name:     "test",
		RolesIDs: []int{1000, 1002, 1003},
		Titles:   []string{"Lead", "Principal"},
		DeptIDs:  []int{10, 13},
	}
	data, err := bintly.Marshal(emp)
	if err != nil {
		log.Fatal(err)
	}
	clone := Employee{}
	err = bintly.Unmarshal(data, &clone)
	if err != nil {
		log.Fatal(err)
	}
}
```

### Custom encoding/decoding

To avoid reflection overhead you can define a custom [encoder](encoder.go) and [decoder](decoder.go) 

#### Struct coder

```go
//EncodeBinary encodes data from binary stream
func (e *Employee) EncodeBinary(stream *bintly.Writer) error {
	stream.Int(e.ID)
	stream.String(e.Name)
	stream.Ints(e.RolesIDs)
	stream.Strings(e.Titles)
	stream.Ints(e.DeptIDs)
	return nil
}

//DecodeBinary decodes data to binary stream
func (e *Employee) DecodeBinary(stream *bintly.Reader) error {
	stream.Int(&e.ID)
	stream.String(&e.Name)
	stream.Ints(&e.RolesIDs)
	stream.Strings(&e.Titles)
	stream.Ints(&e.DeptIDs)
	return nil
}
```

#### Slice coder

```go
type Employees []*Employee

func (e *Employees) DecodeBinary(stream *bintly.Reader) error {
	size := int(stream.Alloc())
	if size == bintly.NilSize {
		return nil
	}
	for i := 0;i< size;i++ {
		if err := stream.Any((*e)[i]);err != nil {
			return err
		}
	}
	return nil
}

func (e *Employees) EncodeBinary(stream *bintly.Writer) error {
	if *e == nil {
		stream.Alloc(bintly.NilSize)
		return nil
	}
	stream.Alloc(int32(len(*e)))
	for i := range *e {
		if err := stream.Any((*e)[i]);err != nil {
			return nil
		}
	}	
	return nil
}

func Example_Slice_Unmarshal() {
    emps := Employees{
		{
			ID:       1,
			Name:     "test 1",
			RolesIDs: []int{1000, 1002, 1003},
			Titles:   []string{"Lead", "Principal"},
			DeptIDs:  []int{10, 13},
		},
		{
			ID:       2,
			Name:     "test 2",
			RolesIDs: []int{1000, 1002, 1003},
			Titles:   []string{"Lead", "Principal"},
			DeptIDs:  []int{10, 13},
		},
	}
	
	data, err := bintly.Marshal(&emps) //pass pointer to the slice
	if err != nil {
		log.Fatal(err)
	}
	var clone Employees
	err = bintly.Unmarshal(data, &clone)
	if err != nil {
		log.Fatal(err)
	}
```

#### Map coder

```go
type EmployeesMap map[int]Employee

func (e *EmployeesMap) DecodeBinary(stream *bintly.Reader) error {
	size := int(stream.Alloc())
	if size == bintly.NilSize {
		return nil
	}
	*e = make(map[int]Employee, size)
	for i := 0; i < size; i++ {
		var k string
		var v Employee
		if err := stream.Any(&k); err != nil {
			return err
		}
		if err := stream.Any(&v); err != nil {
			return err
		}
	}
	return nil
}

func (e *EmployeesMap) EncodeBinary(stream *bintly.Writer) error {
	if *e == nil {
		stream.Alloc(bintly.NilSize)
		return nil
	}
	stream.Alloc(int32(len(*e)))
	for k, v := range *e {
		if err := stream.Any(k); err != nil {
			return nil
		}
		if err := stream.Any(v); err != nil {
			return nil
		}
	}
	return nil
}

func Example_Map_Unmarshal() {
	emps := EmployeesMap{
		1:{
			ID:       1,
			Name:     "test 1",
			RolesIDs: []int{1000, 1002, 1003},
			Titles:   []string{"Lead", "Principal"},
			DeptIDs:  []int{10, 13},
		},
		2:{
			ID:       2,
			Name:     "test 2",
			RolesIDs: []int{1000, 1002, 1003},
			Titles:   []string{"Lead", "Principal"},
			DeptIDs:  []int{10, 13},
		},
	}
	data, err := bintly.Marshal(&emps) //pass pointer to the map
	if err != nil {
		log.Fatal(err)
	}
	var clone EmployeesMap
	err = bintly.Unmarshal(data, &clone)
	if err != nil {
		log.Fatal(err)
	}
}
```
###Auto Code Generation
The package provides CLI options to generate bintly encoding/decoding to make  conversion effortless. For instance , the following command generates encode/decode for
a type "Message" defined in the file "message.go" to produce "message_enc.go" in the same directory
```go
./gen -s "/Users/xxx/github.com/viant/bintly/codegen/test_data/slices/message.go" -t "Message"
```
Options :
```yaml
    -s : Source file location (required)
    -t : Type (required)
    -d : Destination file directory (optional - if not provided output is produced in same directory as source)
```

####Examples
##### Basic Types
Input : message.go
```go
package basic_struct

type Message struct {
	A1 int
	B1 *string
	C1 []string
}
```
Cmd:
```go
gen -s "message.go" -t "Message" 
```
Output : message_enc.go
```go
package basic_struct

import (
	"github.com/viant/bintly"
)
func (m *Message) EncodeBinary(coder *bintly.Writer) error {
	coder.Int(m.A1)
	coder.StringPtr(m.B1)
	coder.Strings(m.C1)
	return nil
}
func (m *Message) DecodeBinary(coder *bintly.Reader) error {
	coder.Int(&m.A1)
	coder.StringPtr(&m.B1)
	coder.Strings(&m.C1)	
	return nil
}

```
##### Slice Types
Input:
```go
package slices

type SubMessage struct {
	Id   int
	Name string
}

type Message struct {
	M1 []SubMessage
}
```
Output :

```go
package slices

import (
	"github.com/viant/bintly"
)

func (s *SubMessage) EncodeBinary(coder *bintly.Writer) error {
	coder.Int(s.Id)
	coder.String(s.Name)
	return nil
}

func (s *SubMessage) DecodeBinary(coder *bintly.Reader) error {
	coder.Int(&s.Id)
	coder.String(&s.Name)	
	return nil
}

func (m *Message) EncodeBinary(coder *bintly.Writer) error {
	var m1 = len(m.M1)
	coder.Alloc(int32(m1))
	for i:=0; i < m1 ; i++ {
		if err := coder.Coder(&m.M1[i]);err !=nil {
			return nil
		}
	}
	return nil
}

func (m *Message) DecodeBinary(coder *bintly.Reader) error {
	var m1 = coder.Alloc()
	m.M1 = make([]SubMessage,m1)
	for i:=0; i < int(m1) ; i++ {
		if err := coder.Coder(&m.M1[i]);err != nil {
			return nil
		}
	}	
	return nil
}
```
##### Map Types
Input:
```go
package maps

type SubMessage struct {
	Id   int
	Name string
}


type M1 map[string][]*SubMessage

type Message struct {
	 M1
}
```
Output :
```go
package maps

import (
	"github.com/viant/bintly"
)

func (s *SubMessage) EncodeBinary(coder *bintly.Writer) error {
	coder.Int(s.Id)
	coder.String(s.Name)
	return nil
}
func (s *SubMessage) DecodeBinary(coder *bintly.Reader) error {
	coder.Int(&s.Id)
	coder.String(&s.Name)
	return nil
}

func (m *Message) EncodeBinary(coder *bintly.Writer) error {
	coder.Alloc(int32(len(m.M1)))
	for k, v := range m.M1 {
		coder.String(k)
		var m1 = len(v)
		coder.Alloc(int32(m1))
		for i := 0; i < m1; i++ {
			if err := coder.Coder(v[i]); err != nil {
				return nil
			}
		}
	}
	return nil
}
func (m *Message) DecodeBinary(coder *bintly.Reader) error {
	size := int(coder.Alloc())
	if size == bintly.NilSize {
		return nil
	}
	m.M1 = make(map[string][]*SubMessage, size)
	for i := 0; i < size; i++ {
		var k string
		var v []*SubMessage
		coder.String(&k)
		var m1Size = coder.Alloc()
		v = make([]*SubMessage, m1Size)
		for j := 0; j < int(m1Size); j++ {
			v[j] = &SubMessage{}
			if err := coder.Coder(v[j]); err != nil {
				return nil
			}
		}
		m.M1[k] = v
	}
	return nil
}

```


### Bugs

This package uses architecture dependent encoding (native go representation).
User of the package should ensure that it is compatible with the needed platforms and architectures.


### Benchmark

Benchmark uses [BenchStruct](stress/bench.go)  where slices got populated with 80 random items.

```bash
BenchmarkUnmarshalBintly
BenchmarkUnmarshalBintly-16           	  940587	      1263 ns/op	    3762 B/op	       6 allocs/op
BenchmarkMarshalBintly
BenchmarkMarshalBintly-16             	  856911	      1202 ns/op	    2484 B/op	       3 allocs/op
BenchmarkUnmarshalBintlyReflect
BenchmarkUnmarshalBintlyReflect-16    	  675956	      1675 ns/op	    3796 B/op	       7 allocs/op
BenchmarkMarshalBintlyReflect
BenchmarkMarshalBintlyReflect-16      	  580586	      1784 ns/op	    2507 B/op	      10 allocs/op
BenchmarkUnmarshalBinary
BenchmarkUnmarshalBinary-16           	  365941	      3306 ns/op	    3152 B/op	      73 allocs/op
BenchmarkMarshalBinary
BenchmarkMarshalBinary-16             	  239450	      4868 ns/op	    4536 B/op	       7 allocs/op
BenchmarkUnMarshalGob
BenchmarkUnMarshalGob-16              	   46789	     25958 ns/op	   13928 B/op	     317 allocs/op
BenchmarkMarshalGob
BenchmarkMarshalGob-16                	   96601	     10510 ns/op	    9956 B/op	      36 allocs/op
BenchmarkUnmarshalCbor
BenchmarkUnmarshalCbor-16             	  109017	     11655 ns/op	    3488 B/op	      79 allocs/op
BenchmarkMarshalCbor
BenchmarkMarshalCbor-16               	  216528	      5453 ns/op	    2194 B/op	       2 allocs/op
BenchmarkUnmarshalMsgPack
BenchmarkUnmarshalMsgPack-16          	   69460	     17462 ns/op	    4899 B/op	      84 allocs/op
BenchmarkMarshalMsgPack
BenchmarkMarshalMsgPack-16            	   99207	     12255 ns/op	    4722 B/op	       8 allocs/op
BenchmarkJSONUnmarshal
BenchmarkJSONUnmarshal-16             	   18126	     72414 ns/op	   15488 B/op	     308 allocs/op
BenchmarkJSONMarshal
BenchmarkJSONMarshal-16               	   52039	     21745 ns/op	    4359 B/op	       3 allocs/op
```



<a name="License"></a>
## License

The source code is made available under the terms of the Apache License, Version 2, as stated in the file `LICENSE`.

Individual files may be made available under their own specific license,
all compatible with Apache License, Version 2. Please see individual files for details.

<a name="Credits-and-Acknowledgements"></a>

## Contributing to Bintly

Bintly is an open source project and contributors are welcome!

See [TODO](TODO.md) list

## Credits and Acknowledgements

**Library Author:** Adrian Witas

