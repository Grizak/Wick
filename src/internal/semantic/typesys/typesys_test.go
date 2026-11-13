package typesys_test

import (
	"testing"

	"github.com/Grizak/Wick/src/internal/semantic/typesys"
	"github.com/Grizak/Wick/src/internal/types"
)

func TestParseTypeString(t *testing.T) {
	checker := typesys.NewTypeChecker()
	tests := []struct {
		input string
		name  string
		llvm  string
		size  int
	}{
		{input: "i32", name: "i32", llvm: "i32", size: 4},
		{input: "i64", name: "int", llvm: "i64", size: 8},
		{input: "f64", name: "float", llvm: "double", size: 8},
		{input: "bool", name: "bool", llvm: "i1", size: 1},
		{input: "string", name: "string", llvm: "i8*", size: 8},
		{input: "*i32", name: "*i32", llvm: "i32*", size: 8},
		{input: "[3]i32", name: "[3]i32", llvm: "[3 x i32]", size: 12},
		{input: "[2]*i64", name: "[2]*int", llvm: "[2 x i64*]", size: 16},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := checker.ParseTypeString(tt.input)
			if err != nil {
				t.Fatalf("ParseTypeString(%q): %v", tt.input, err)
			}
			if got.Name() != tt.name || got.LLVMType() != tt.llvm || got.SizeBytes() != tt.size {
				t.Errorf("type = (%q, %q, %d), want (%q, %q, %d)",
					got.Name(), got.LLVMType(), got.SizeBytes(), tt.name, tt.llvm, tt.size)
			}
		})
	}

	for _, invalid := range []string{"unknown", "[x]i32", "[3i32", "*"} {
		if _, err := checker.ParseTypeString(invalid); err == nil {
			t.Errorf("ParseTypeString(%q) succeeded, want error", invalid)
		}
	}
}

func TestTypeEnvironmentScopesAndCoercion(t *testing.T) {
	parent := typesys.NewTypeEnvironment()
	parent.Define("value", &typesys.Int32Type{})
	child := parent.Child()

	if got, ok := child.Lookup("value"); !ok || !got.Equals(&typesys.Int32Type{}) {
		t.Fatalf("child did not inherit parent value: (%v, %v)", got, ok)
	}
	child.Define("value", &typesys.FloatType{})
	if got, ok := child.Lookup("value"); !ok || !got.Equals(&typesys.FloatType{}) {
		t.Fatalf("child shadowing failed: (%v, %v)", got, ok)
	}
	if got, ok := parent.Lookup("value"); !ok || !got.Equals(&typesys.Int32Type{}) {
		t.Fatalf("child definition changed parent value: (%v, %v)", got, ok)
	}
	if _, ok := parent.Lookup("missing"); ok {
		t.Fatal("lookup found an undefined name")
	}

	if !typesys.TryCoerce(&typesys.Int32Type{}, &typesys.FloatType{}) {
		t.Error("numeric coercion rejected")
	}
	if typesys.TryCoerce(&typesys.StringType{}, &typesys.Int32Type{}) {
		t.Error("incompatible coercion accepted")
	}
	if !typesys.TryCoerce(typesys.VoidType{}, &typesys.VoidType{}) {
		t.Error("equal types rejected")
	}

	var _ types.Type = &typesys.ArrayType{ElementType: &typesys.Int32Type{}, Length: 2}
}
