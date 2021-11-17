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

// Config contains all website configuration options for a Monastery website
type Config struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	ContentPath string `json:"content_path"`
	StaticPath  string `json:"static_path"`

	Pinned map[string]string `json:"pinned"`

	Style string `json:"default_style"`

	ScanInterval int `json:"scan_interval"`
}
