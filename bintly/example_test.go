package bintly_test

import (
	"log"

	"github.com/viant/bintly"
)

// Employee represents a test struct
type Employee struct {
	ID       int
	Name     string
	RolesIDs []int
	Titles   []string
	DeptIDs  []int
}

func ExampleUnmarshal() {
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

// DecodeBinary decodes data to binary stream
func (e *Employee) DecodeBinary(stream *bintly.Reader) error {
	stream.Int(&e.ID)
	stream.String(&e.Name)
	stream.Ints(&e.RolesIDs)
	stream.Strings(&e.Titles)
	stream.Ints(&e.DeptIDs)
	return nil
}

// EncodeBinary encodes data from binary stream
func (e *Employee) EncodeBinary(stream *bintly.Writer) error {
	stream.Int(e.ID)
	stream.String(e.Name)
	stream.Ints(e.RolesIDs)
	stream.Strings(e.Titles)
	stream.Ints(e.DeptIDs)
	return nil
}

func ExampleEmployees() {
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

	data, err := bintly.Marshal(&emps) // pass pointer to the slice
	if err != nil {
		log.Fatal(err)
	}
	var clone Employees
	err = bintly.Unmarshal(data, &clone)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleEmployee() {
	emps := EmployeesMap{
		1: {
			ID: 1,

			Name:     "test 1",
			RolesIDs: []int{1000, 1002, 1003},
			Titles:   []string{"Lead", "Principal"},
			DeptIDs:  []int{10, 13},
		},
		2: {
			ID:       2,
			Name:     "test 2",
			RolesIDs: []int{1000, 1002, 1003},
			Titles:   []string{"Lead", "Principal"},
			DeptIDs:  []int{10, 13},
		},
	}

	data, err := bintly.Marshal(&emps) // pass pointer to the map
	if err != nil {
		log.Fatal(err)
	}
	var clone EmployeesMap
	err = bintly.Unmarshal(data, &clone)
	if err != nil {
		log.Fatal(err)
	}
}

// Employees represents Employee slice
type Employees []*Employee

// DecodeBinary decodes data to binary stream
func (e *Employees) DecodeBinary(stream *bintly.Reader) error {
	size := int(stream.Alloc())
	if size == bintly.NilSize {
		return nil
	}
	*e = make([]*Employee, size)
	for i := 0; i < size; i++ {
		if err := stream.Any((*e)[i]); err != nil {
			return err
		}
	}
	return nil
}

// EncodeBinary encodes data from binary stream
func (e *Employees) EncodeBinary(stream *bintly.Writer) error {
	if *e == nil {
		stream.Alloc(bintly.NilSize)
		return nil
	}
	stream.Alloc(int32(len(*e)))
	for i := range *e {
		if err := stream.Any((*e)[i]); err != nil {
			return nil
		}
	}
	return nil
}

// EmployeesMap represents employee maps
type EmployeesMap map[int]Employee

// DecodeBinary decodes data to binary stream
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

// EncodeBinary encodes data from binary stream
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
