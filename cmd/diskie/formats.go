package main

var formatTabular = `
{{
	printf "%-20s   %7s   %-12s   %-15s   %s"
	( .DriveModel | default "-" | condense | abbrev 20 )
	( .PreferredSize | humanBytesIEC )
	( .IdType  | dereference | default "-" | abbrev 12 )
	( .IdLabel | dereference | default "-" | abbrev 15 )
	( .Device  | dereference )
}}
`

var formatBasic = `
{{
	$vars := list
		( .DriveModel | condense )
		( .PreferredSize | humanBytesIEC )
		( .IdType )
		( .IdLabel )
		( .Device )
}}
{{ range $vars }}
{{ if not (isEmpty .) }}
[ {{ . | abbrev 20 }} ]
{{ " " }}
{{ end }}
{{ end }}
`

var formatRofiMarkup = `
{{
	$vars := list
		( .DriveModel | condense )
		( .PreferredSize | humanBytesIEC )
		( .IdType )
		( .IdLabel )
		( .Device )
}}
{{ range $vars }}
{{ if not (isEmpty .) }}
<span alpha="50%" weight="100">[ </span>{{ . | abbrev 20 }}<span alpha="50%" weight="100"> ]</span>
{{ " " }}
{{ end }}
{{ end }}
`
