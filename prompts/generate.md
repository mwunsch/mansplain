Output a complete mdoc(7) man page. No markdown, no explanation, just the mdoc source.

Follow this example exactly:

.Dd March 26, 2026
.Dt GREP 1
.Os
.Sh NAME
.Nm grep
.Nd file pattern searcher
.Sh SYNOPSIS
.Nm
.Op Fl c
.Op Fl i
.Op Fl n
.Op Fl r
.Op Fl v
.Op Fl E
.Op Fl F
.Op Fl A Ar num
.Op Fl B Ar num
.Op Fl C Ar num
.Op Fl e Ar pattern
.Op Fl f Ar file
.Op Ar pattern
.Op Ar
.Sh DESCRIPTION
The
.Nm
utility searches input files for lines matching a pattern.
By default, matching lines are printed to standard output.
.Sh OPTIONS
.Bl -tag -width indent
.It Fl c , Fl -count
Print only a count of matching lines.
.It Fl i , Fl -ignore-case
Case insensitive matching.
.It Fl n , Fl -line-number
Prefix each line with its line number.
.It Fl r , Fl -recursive
Recursively search directories.
.It Fl v , Fl -invert-match
Select non-matching lines.
.It Fl E , Fl -extended-regexp
Use extended regular expressions.
.It Fl A Ar num
Print
.Ar num
lines after each match.
.It Fl B Ar num
Print
.Ar num
lines before each match.
.It Fl C Ar num
Print
.Ar num
lines of context around each match.
.It Fl e Ar pattern
Specify a search pattern.
.El
.Sh ENVIRONMENT
.Bl -tag -width indent
.It Ev GREP_OPTIONS
Default options prepended to the argument list.
.El
.Sh EXIT STATUS
.Ex -std
.Sh EXAMPLES
Search for a pattern in a file:
.Bd -literal -offset indent
grep 'error' /var/log/syslog
.Ed
.Pp
Recursive case-insensitive search:
.Bd -literal -offset indent
grep -ri 'todo' src/
.Ed
.Pp
Show context around matches:
.Bd -literal -offset indent
grep -C 3 'panic' *.go
.Ed
.Sh SEE ALSO
.Xr awk 1 ,
.Xr sed 1
.Sh STANDARDS
The
.Nm
utility is compliant with
.St -p1003.1-2008 .

Rules:
- .Dt title must be UPPERCASE
- .Os takes no argument
- Every flag from the source material gets an .It entry in OPTIONS
- Use .Op for optional args in SYNOPSIS, .Fl for flags, .Ar for arguments
- Use .Cm for subcommand names
- Use .Bl -tag -width indent / .It / .El for option lists
- Use .Bd -literal -offset indent / .Ed for example code blocks
- Use .Ev for environment variables, .Pa for paths, .Xr for cross-references
- Never use raw troff (.ft, .sp, .in, .br, .nf, .fi, .RS, .RE)
- Include EXAMPLES with 2-3 real invocations
- Be terse. Man pages are reference, not tutorials.
