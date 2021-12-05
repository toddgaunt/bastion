package document

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

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
