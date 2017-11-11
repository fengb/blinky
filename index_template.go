/*
 * CODE GENERATED AUTOMATICALLY WITH
 *    github.com/wlbr/templify
 * THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package main

// index_templateTemplate is a generated function returning the template as a string.
// That string should be parsed by the functions of the golang's template package.
func index_templateTemplate() string {
	var tmpl = "<html>\n" +
		"  <h3>Last Synced &mdash; {{ .LastSync }}</h3>\n" +
		"  <table>\n" +
		"          <tr>\n" +
		"                  <th>Name</th>\n" +
		"                  <th>Version</th>\n" +
		"                  <th>Upgrade</th>\n" +
		"          </tr>\n" +
		"          {{- range .Packages }}\n" +
		"                  <tr>\n" +
		"                          <td>{{ .Name }}</td>\n" +
		"                          <td>{{ .Version }}</td>\n" +
		"                          <td>{{ .Upgrade }}</td>\n" +
		"                  </tr>\n" +
		"          {{- end }}\n" +
		"  </table>\n" +
		"</html>\n" +
		""
	return tmpl
}
