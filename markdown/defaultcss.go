package markdown

// defaultCSS provides a basic stylesheet for Markdown-generated HTML.
const defaultCSS = `
@page {
	size: a4;
	margin: 2cm;
}

body {
	font-family: serif;
	font-size: 10pt;
	line-height: 1.4;
}

h1 {
	font-size: 24pt;
	margin-top: 18pt;
	margin-bottom: 12pt;
	font-weight: bold;
}

h2 {
	font-size: 18pt;
	margin-top: 14pt;
	margin-bottom: 10pt;
	font-weight: bold;
}

h3 {
	font-size: 14pt;
	margin-top: 12pt;
	margin-bottom: 8pt;
	font-weight: bold;
}

h4 {
	font-size: 12pt;
	margin-top: 10pt;
	margin-bottom: 6pt;
	font-weight: bold;
}

h5 {
	font-size: 10pt;
	margin-top: 8pt;
	margin-bottom: 4pt;
	font-weight: bold;
}

h6 {
	font-size: 10pt;
	margin-top: 6pt;
	margin-bottom: 4pt;
	font-weight: bold;
	font-style: italic;
}

h1, h2, h3, h4, h5, h6 {
	break-after: avoid;
}

p {
	margin-top: 6pt;
	margin-bottom: 6pt;
}

blockquote {
	margin-left: 20pt;
	margin-right: 20pt;
	margin-top: 6pt;
	margin-bottom: 6pt;
	font-style: italic;
}

pre {
	font-family: monospace;
	font-size: 9pt;
	margin-top: 6pt;
	margin-bottom: 6pt;
	padding: 6pt;
	background-color: #f5f5f5;
}

code {
	font-family: monospace;
	font-size: 9pt;
}

table {
	margin-top: 8pt;
	margin-bottom: 8pt;
	border-collapse: collapse;
}

th {
	font-weight: bold;
	padding: 2pt 6pt;
	border-bottom: 1pt solid black;
	line-height: 1.2;
}

td {
	padding: 2pt 6pt;
	line-height: 1.2;
}

ul, ol {
	margin-top: 4pt;
	margin-bottom: 4pt;
	padding-left: 20pt;
}

li {
	margin-top: 2pt;
	margin-bottom: 2pt;
}

hr {
	margin-top: 12pt;
	margin-bottom: 12pt;
	border-top: 0.5pt solid #999999;
}

strong, b {
	font-weight: bold;
}

em, i {
	font-style: italic;
}
`
