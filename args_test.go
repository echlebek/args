package args

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestRawArgsMap(t *testing.T) {
	args := []string{
		"--foo", "asdf", "asdf",
	}

	got, err := rawArgsMap(args)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string][]string{
		"foo": []string{"asdf", "asdf"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestSlice(t *testing.T) {
	args := []string{
		"--foo",
		"foobar",
		"--bar",
		"barbar",
		"--baz",
	}

	var stringData []string
	if err := parse(stringData, args); err == nil {
		t.Fatal("expected error")
	}
	if err := parse(&stringData, args); err != nil {
		t.Fatal(err)
	}
	if want := []string{"--foo", "foobar", "--bar", "barbar", "--baz"}; !reflect.DeepEqual(stringData, want) {
		t.Fatalf("bad data: got %v, want %v", stringData, want)
	}
	args = []string{
		"--foo",
		"5",
		"--bar",
		"10",
	}
	var intData []int
	var uintData []uint
	if err := parse(&intData, args); err == nil {
		t.Fatal("expected error")
	}
	if err := parse(&uintData, args); err == nil {
		t.Fatal("expected error")
	}
	args = []string{"5", "10"}
	if err := parse(&intData, args); err != nil {
		t.Fatal(err)
	}
	if want := []int{5, 10}; !reflect.DeepEqual(want, intData) {
		t.Fatalf("bad data: got %v, want %v", intData, want)
	}
	if err := parse(&uintData, args); err != nil {
		t.Fatal(err)
	}
	if want := []uint{5, 10}; !reflect.DeepEqual(want, uintData) {
		t.Fatalf("bad data: got %v, want %v", intData, want)
	}
	var f32Data []float32
	var f64Data []float64
	args = []string{"abc", "-0.123", "1.1"}
	if err := parse(&f32Data, args); err == nil {
		t.Fatal("expected error")
	}
	if err := parse(&f32Data, args[1:]); err != nil {
		t.Fatal(err)
	}
	if want := []float32{-0.123, 1.1}; !reflect.DeepEqual(want, f32Data) {
		t.Fatalf("bad data: got %v, want %v", f32Data, want)
	}
	if err := parse(&f64Data, args[1:]); err != nil {
		t.Fatal(err)
	}
	if want := []float64{-0.123, 1.1}; !reflect.DeepEqual(want, f64Data) {
		t.Fatalf("bad data: got %v, want %v", f64Data, want)
	}
}

func TestMap(t *testing.T) {
	args := []string{
		"--foo",
		"foo",
		"--bar",
		"bar",
		"--baz",
		"baz",
		"--asdf",
	}
	badMap := make(map[int]int)
	if err := parse(&badMap, args); err == nil {
		t.Fatal("expected error")
	}
	strMap := make(map[string]string)
	if err := parse(&strMap, args); err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"foo": "foo", "bar": "bar", "baz": "baz", "asdf": ""}; !reflect.DeepEqual(want, strMap) {
		t.Fatalf("bad data: got %+v, want %+v", strMap, want)
	}
	args = append(args, "--foo", "bad")
	if err := parse(&strMap, args); err == nil {
		t.Fatal("expected error")
	}
	args = []string{"--foo", "5", "--bar", "10.5", "--baz"}
	intMap := make(map[string]int)
	if err := parse(&intMap, args); err == nil {
		t.Fatal("expected error")
	}
	ifaceMap := make(map[string]interface{})
	if err := parse(&ifaceMap, args); err != nil {
		t.Fatal(err)
	}
	if want := map[string]interface{}{"foo": int64(5), "bar": float64(10.5), "baz": struct{}{}}; !reflect.DeepEqual(ifaceMap, want) {
		t.Fatalf("bad data: got %+v, want %+v", ifaceMap, want)
	}
}

func TestStruct(t *testing.T) {
	type Test struct {
		Int8    int8 `args:"this is an int,r,3"`
		Int16   int16
		Int32   int32
		Int64   int64
		Int     int
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Uint    uint
		Float32 float32
		Float64 float64
		String  string
		Bool    bool `args:"a switch,-b"`
		Slice   []string

		Int8ptr    *int8 `args:"foo bar"`
		Int16ptr   *int16
		Int32ptr   *int32
		Int64ptr   *int64
		Intptr     *int
		Uint8ptr   *uint8
		Uint16ptr  *uint16
		Uint32ptr  *uint32
		Uint64ptr  *uint64
		Uintptr    *uint
		Float32ptr *float32
		Float64ptr *float64
		Stringptr  *string
		Boolptr    *bool `args:"another switch,-p"`
		Sliceptr   *[]string

		unexported int // make sure we don't try to set an unexported field
	}

	var got Test

	testArgs := []string{
		"--int8", "-1",
		"--int16", "2",
		"--int32", "-3",
		"--int64", "4",
		"--int", "5",
		"--uint8", "1",
		"--uint16", "2",
		"--uint32", "3",
		"--uint64", "4",
		"--uint", "5",
		"--float32", "1.2",
		"--float64", "1.1",
		"--string", "foo",
		"--bool",
		"--slice", "a", "b",
		"--int8ptr", "-1",
		"--int16ptr", "2",
		"--int32ptr", "-3",
		"--int64ptr", "4",
		"--intptr", "5",
		"--uint8ptr", "1",
		"--uint16ptr", "2",
		"--uint32ptr", "3",
		"--uint64ptr", "4",
		"--uintptr", "5",
		"--float32ptr", "1.2",
		"--float64ptr", "1.1",
		"--stringptr", "foo",
		"--boolptr",
		"--sliceptr", "a", "b",
	}

	if err := parse(&got, testArgs); err != nil {
		t.Fatal(err)
	}

	var (
		i8    int8     = -1
		i16   int16    = 2
		i32   int32    = -3
		i64   int64    = 4
		i     int      = 5
		u8    uint8    = 1
		u16   uint16   = 2
		u32   uint32   = 3
		u64   uint64   = 4
		u     uint     = 5
		f32   float32  = 1.2
		f64   float64  = 1.1
		s     string   = "foo"
		b     bool     = true
		slice []string = []string{"a", "b"}
	)

	want := Test{
		Int8:       i8,
		Int16:      i16,
		Int32:      i32,
		Int64:      i64,
		Int:        i,
		Uint8:      u8,
		Uint16:     u16,
		Uint32:     u32,
		Uint64:     u64,
		Uint:       u,
		Float32:    f32,
		Float64:    f64,
		String:     s,
		Bool:       b,
		Slice:      slice,
		Int8ptr:    &i8,
		Int16ptr:   &i16,
		Int32ptr:   &i32,
		Int64ptr:   &i64,
		Intptr:     &i,
		Uint8ptr:   &u8,
		Uint16ptr:  &u16,
		Uint32ptr:  &u32,
		Uint64ptr:  &u64,
		Uintptr:    &u,
		Float32ptr: &f32,
		Float64ptr: &f64,
		Stringptr:  &s,
		Boolptr:    &b,
		Sliceptr:   &slice,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bad data: got %+v, want %+v", got, want)
	}

	// Remove a required arg
	testArgs = testArgs[1:]
	if err := parse(&got, testArgs); err == nil {
		t.Fatal("expected error")
	}

}

type UsageTest struct {
	Salad  string   `args:"type of salad to eat,-s"`
	Pie    int      `args:"number of pies to eat,-p"`
	Nachos *float32 `args:"nacho quotient"`
}

func (t UsageTest) Describe(w io.Writer) error {
	_, err := fmt.Fprint(w, "Foods to eat: a mock program")
	return err
}

func TestUsage(t *testing.T) {
	if err := Usage(nil, 1); err == nil {
		t.Fatal("expected error")
	}

	var ts UsageTest

	var buf bytes.Buffer

	if err := Usage(&buf, ts); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"usage:",
		" -s,\t--salad\t(default: \"\")\ttype of salad to eat",
		" -p,\t--pie\t(default: 0)\tnumber of pies to eat",
		" \t--nachos\t\tnacho quotient",
		"", // A result of the splitting
	}

	for i, got := 0, strings.Split(buf.String(), "\n"); i < len(got); i++ {
		if len(got) != len(want) {
			t.Fatalf("got %d lines, want %d lines", len(got), len(want))
		}
		if got[i] != want[i] {
			t.Fatalf("bad usage on line %d: got %s, want %s", i, got[i], want[i])
		}
	}
}

func TestCheckArgLen(t *testing.T) {
	a := []string{}
	b := []string{"a"}
	c := []string{"a", "b"}
	if err := checkArgLen(b, "b"); err != nil {
		t.Fatal(err)
	}
	if err := checkArgLen(a, "a"); err == nil {
		t.Fatal("expected error")
	}
	if err := checkArgLen(c, "c"); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetType(t *testing.T) {
	type Foo struct{}
	var f Foo
	if _, err := getType(f); err == nil {
		t.Fatal("expected error")
	}
	if typ, err := getType(&f); err != nil {
		t.Fatal(err)
	} else if typ.Kind() != reflect.Struct {
		t.Fatalf("expected reflect.Struct, got %s", typ.Kind().String())
	}
	var s []string
	if typ, err := getType(&s); err != nil {
		t.Fatal(err)
	} else if typ.Kind() != reflect.Slice {
		t.Fatalf("expected reflect.Slice, got %s", typ.Kind().String())
	}
	var m map[string]string
	if typ, err := getType(&m); err != nil {
		t.Fatal(err)
	} else if typ.Kind() != reflect.Map {
		t.Fatalf("expected reflect.Map, got %s", typ.Kind().String())
	}
	var x int
	if _, err := getType(&x); err == nil {
		t.Fatal("expected error")
	}
}
