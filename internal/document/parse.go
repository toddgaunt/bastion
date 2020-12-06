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

package document

import (
	"bufio"
	"bytes"
	"errors"
	"regexp"
	"strings"
)

type Document struct {
	Properties Properties
	Format     string
	Content    []byte
}

type Properties map[string][]string

// Add adds a key and value to a property
func (p Properties) Add(key, value string) {
	values, ok := p[strings.ToLower(key)]
	if !ok {
		p[key] = []string{value}
	}

	p[key] = append(values, value)
}

// Value returns the first value associated with a key
func (p Properties) Value(key string) string {
	values, ok := p[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return values[0]
}

// Values returns all of the values associated with a key
func (p Properties) Values(key string) []string {
	values, ok := p[strings.ToLower(key)]
	if !ok {
		return nil
	}
	return values
}

// Parse parses bytes and returns a Document, or an error if the bytes did not
// form a valid document representation
func Parse(data []byte) (Document, error) {
	re := regexp.MustCompile(`===.*===`)
	index := re.FindIndex(data)
	if index == nil {
		return Document{}, errors.New("document does not have article delimiter")
	}

	properties, err := parseProperties(data[:index[0]])
	if err != nil {
		return Document{}, err
	}
	format := strings.TrimSpace(string(data[index[0]+3 : index[1]-3]))
	content := data[index[1]:]

	return Document{
		Properties: properties,
		Format:     format,
		Content:    content,
	}, nil
}

func parseProperties(data []byte) (Properties, error) {
	properties := make(Properties)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		text := scanner.Text()
		// Ignore blank lines
		if strings.TrimSpace(text) == "" {
			continue
		}

		// KEY : VALUE syntax is expected on non blank lines
		split := strings.SplitN(text, ":", 2)
		if len(split) != 2 {
			return nil, errors.New("expected 'KEY : VALUE' pair")
		}

		key := strings.ToLower(strings.TrimSpace(split[0]))
		value := strings.TrimSpace(split[1])

		properties.Add(key, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return properties, nil
}
