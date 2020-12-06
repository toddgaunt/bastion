// Copyright 2020, Todd Gaunt <toddgaunt@protonmail.com>
//
// This file is part of Monastery.
//
// Monastery is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Monastery is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Monastery.  If not, see <https://www.gnu.org/licenses/>.

package monastery

const siteHeaderHTML = `<!DOCTYPE html>
<html>
<head>
  <title>{{if .Config.Title}}{{.Config.Title}}{{else}}Site{{end}}</title>
  <meta name="description" content="{{.Config.Description}}">
  {{if .Config.Style}}<link href="/.static/styles/{{.Config.Style}}.css" type="text/css" rel="stylesheet">{{end}}
</head>
<body>
<div id="site-header">
	<h1 id="site-title">{{if .Config.Title}}{{.Config.Title}}{{else}}Site{{end}}</h1>
	<div id="site-navigation">
	<a href="/">Articles</a>{{range $k, $v := .Config.Pinned}}
	<a href="/{{$v}}">{{$k}}</a>{{end}}
	</div>
</div>
<div id="content">
`

const siteFooterHTML = `</div>
</body>
</html>
`
