package checker

import "testing"

func TestParseControllerCallbackFieldWithNestedParens(t *testing.T) {
	t.Parallel()

	stmt := "MaaBool (*request_param)(MaaStringBuffer out, MaaBool (*accept)(int, int (*next)(void)), void* trans_arg)"
	retType, name, paramsRaw, ok := parseControllerCallbackField(stmt)
	if !ok {
		t.Fatalf("expected callback field to parse")
	}
	if retType != "MaaBool" {
		t.Fatalf("unexpected return type: %s", retType)
	}
	if name != "request_param" {
		t.Fatalf("unexpected callback name: %s", name)
	}
	want := "MaaStringBuffer out, MaaBool (*accept)(int, int (*next)(void)), void* trans_arg"
	if paramsRaw != want {
		t.Fatalf("unexpected params: got=%q want=%q", paramsRaw, want)
	}
}

func TestParseControllerCallbackFields(t *testing.T) {
	t.Parallel()

	block := `
	/* callback block */
	MaaBool (*request_uuid)(MaaStringBuffer out, void* trans_arg);
	MaaBool (*do_something)(
		MaaBool (*cb)(int, int (*next)(void)),
		void* trans_arg
	);
`
	fields := parseControllerCallbackFields(block)
	if len(fields) != 2 {
		t.Fatalf("unexpected callback count: %d", len(fields))
	}
	if fields[0].name != "request_uuid" {
		t.Fatalf("unexpected first callback: %s", fields[0].name)
	}
	if fields[1].name != "do_something" {
		t.Fatalf("unexpected second callback: %s", fields[1].name)
	}
}
