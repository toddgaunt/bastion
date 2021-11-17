// Copyright 2021, Todd Gaunt <toddgaunt@protonmail.com>
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

// SiteConfig contains all configuration for a Monastery website's commonindex and
// content pages, such as website title, css style, and which articles are
// pinned to the navigation bar instead of being indexed.
type SiteConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Style       string            `json:"style"`
	Pinned      map[string]string `json:"pinned"`
}
