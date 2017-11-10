/*
 * CODE GENERATED AUTOMATICALLY WITH
 *    github.com/wlbr/templify
 * THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package main

// web_indexTemplate is a generated function returning the template as a string.
// That string should be parsed by the functions of the golang's template package.
func web_indexTemplate() string {
	var tmpl = "<table>\n" +
		"\t<tr>\n" +
		"\t\t<th>Name</th>\n" +
		"\t\t<th>Version</th>\n" +
		"\t\t<th>Upgrade</th>\n" +
		"\t</tr>\n" +
		"\t{{- range .}}\n" +
		"\t\t<tr>\n" +
		"\t\t\t<td>{{ .Name }}</td>\n" +
		"\t\t\t<td>{{ .Version }}</td>\n" +
		"\t\t\t<td>{{ .Upgrade }}</td>\n" +
		"\t\t</tr>\n" +
		"\t{{- end }}\n" +
		"</table>\n" +
		""
	return tmpl
}
