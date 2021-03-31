package fpath

import (
	"fmt"
	"reflect"
	"testing"
)

func TestJoin(t *testing.T) {
	type args struct {
		elem []string
	}
	tests := []struct {
		name string
		args args
		want Path
	}{
		{"empty", args{[]string{"", ""}}, Path("")},
		{"/", args{[]string{"", "/"}}, Path("/")},
		{"../a/b", args{[]string{"..", "a", "b"}}, Path("../a/b")},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Join(tt.args.elem...); *got != tt.want {
				t.Errorf("Join() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want Path
	}{
		{"simple", args{"a/b"}, Path("a/b")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.path); *got != tt.want {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Join(t *testing.T) {
	type args struct {
		elem []string
	}
	tests := []struct {
		name string
		p    Path
		args args
		want Path
	}{
		{"empty", Path(""), args{[]string{"", ""}}, Path("")},
		{"a/b/c", Path("a"), args{[]string{"b", "c"}}, Path("a/b/c")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Join(tt.args.elem...); *got != tt.want {
				t.Errorf("Join() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Abs(t *testing.T) {
	tests := []struct {
		name string
		p    Path
		want Path
	}{
		{"/a/b/c", Path("/a/../../x"), Path("/x")},
		{"Cwd/a/b/c", Path("a/../x"), *Cwd().Join("/x")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Abs(); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("Abs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPath_Parents(t *testing.T) {
	type args struct {
		level int
	}
	tests := []struct {
		name string
		p    Path
		args args
		want Path
	}{
		{":empty", Path(""), args{ 0}, Path(".")},
		{":/", Path("/"), args{ 0}, Path("/")},
		{":/a", Path("/a"), args{ 0}, Path("/")},
		{":/a/b", Path("/a/b"), args{ 1}, Path("/")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Parents(tt.args.level); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("Parents() = %v, want %v", got, tt.want)
			} else {
				fmt.Println(got)
			}
		})
	}
}

func TestPath_Name(t *testing.T) {
	p := New("$HOME/t.js").Expand()
	fmt.Println(p)
	fmt.Println(p.Dir())
	fmt.Println(p.Base())
	fmt.Println(p.Stem())
	fmt.Println(p.Ext())
	fmt.Println(p.WithPrefix("new-"))
	fmt.Println(p.WithSuffix(".gz"))
}

func TestValue_Array(t *testing.T) {
	type args struct {
		sep string
	}
	tests := []struct {
		name string
		v    Value
		args args
		want []Value
	}{
		{"single: 1", Value("1"), args{" "}, []Value{Value("1")}},
		{"multiple: 1 2", Value("1 2"), args{" "} , []Value{Value("1"), Value("2")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Array(tt.args.sep); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Array() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValue_LoadValueMap(t *testing.T) {
	inp,_ := New("data.txt").Open(ForReading)
	kv := LoadValueMap(inp, ":", true, true, true)

	for k,v := range kv {
		fmt.Println(k, "=", v.Int())
	}
	fmt.Println(kv["User"].Path().ListDir(true))
}