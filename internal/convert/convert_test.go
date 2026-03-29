package convert

import (
	"strings"
	"testing"
)

func TestConvert_TitleLine(t *testing.T) {
	input := []byte("# mytool(1) -- a great tool\n\n## DESCRIPTION\n\nDoes things.\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".Dd $Mdocdate$")
	assertContains(t, s, ".Dt MYTOOL 1")
	assertContains(t, s, ".Os")
	assertContains(t, s, ".Nm mytool")
	assertContains(t, s, ".Nd a great tool")
}

func TestConvert_Sections(t *testing.T) {
	input := []byte("# tool(1) -- desc\n\n## SYNOPSIS\n\n`tool`\n\n## DESCRIPTION\n\nText.\n\n## OPTIONS\n\nNone.\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".Sh SYNOPSIS")
	assertContains(t, s, ".Sh DESCRIPTION")
	assertContains(t, s, ".Sh OPTIONS")
}

func TestConvert_DefinitionList(t *testing.T) {
	input := []byte("# t(1) -- d\n\n## OPTIONS\n\n  * `-v`, `--verbose`:\n    Be loud.\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".Bl -tag -width indent")
	assertContains(t, s, ".It Fl v")
	assertContains(t, s, "Fl -verbose")
	assertContains(t, s, "Be loud.")
	assertContains(t, s, ".El")
}

func TestConvert_CodeBlock(t *testing.T) {
	input := []byte("# t(1) -- d\n\n## EXAMPLES\n\nDo this:\n\n    cmd --flag arg\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".Bd -literal -offset indent")
	assertContains(t, s, "cmd --flag arg")
	assertContains(t, s, ".Ed")
}

func TestConvert_SeeAlso(t *testing.T) {
	input := []byte("# t(1) -- d\n\n## SEE ALSO\n\nfoo(1), bar(3)\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".Xr foo 1 ,")
	assertContains(t, s, ".Xr bar 3")
}

func TestConvert_AngleBrackets(t *testing.T) {
	input := []byte("# t(1) -- d\n\n## DESCRIPTION\n\nTakes a <file> argument.\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	// <file> should become italic in mdoc
	assertContains(t, s, "\\fIfile\\fR")
}

func TestConvert_EnvVar(t *testing.T) {
	input := []byte("# t(1) -- d\n\n## ENVIRONMENT\n\n  * `MY_VAR`:\n    A config var.\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".It Ev MY_VAR")
}

func TestConvert_PlainH1(t *testing.T) {
	// H1 without ronn title format
	input := []byte("# My Tool\n\n## DESCRIPTION\n\nDoes stuff.\n")
	out, err := Convert(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContains(t, s, ".Dt MY TOOL 1")
	assertContains(t, s, ".Nm My Tool")
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("output missing %q\n\nFull output:\n%s", substr, s)
	}
}
