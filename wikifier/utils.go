package wikifier

import (
	"fmt"
	"strings"
)

func pageNameLink(s string) string {
	return ""
}

// convert wikifier values to human-readable form
func humanReadableValue(i interface{}) string {
	switch v := i.(type) {

	// FIXME: recursions possible?

	// list
	case []interface{}:
		hrValues := make([]string, len(v))
		for i, val := range v {
			hrValues[i] = humanReadableValue(val)
		}
		return strings.Join(hrValues, " ")

	// nothing
	case nil:
		return ""

	// string
	case string:
		return `'` + strings.TrimSpace(v) + `'`

	// boolean
	case bool:
		if v {
			return "true"
		}
		return "false"

	// block
	case block:
		return v.String()

		// something else
	default:
		return fmt.Sprintf("%v", v)
	}
}

// fix a value before storing it in a list or map
// this returns either a string, block, or []interface{} of both
// strings next to each other are merged; empty strings are removed
func fixValuesForStorage(values []interface{}) interface{} {

	// no items
	if len(values) == 0 {
		return nil
	}

	// one value in; one value out!
	if len(values) == 1 {
		return fixSingleValue(values[0])
	}

	// multiple values
	var valuesToStore []interface{}
	var lastValue interface{}
	for _, value := range values {

		// fix this value; then skip it if it's nothin
		value = fixSingleValue(value)
		if value == nil {
			continue
		}

		// if this is a string and the previous one was too, combine them
		thisStr, isStr := value.(string)
		lastStr, lastWasStr := lastValue.(string)
		if isStr && lastWasStr {
			valuesToStore[len(valuesToStore)-1] = lastStr + thisStr
		} else {
			valuesToStore = append(valuesToStore, value)
		}

		lastValue = value
	}

	// ended up with nothing
	if len(valuesToStore) == 0 {
		return nil
	}

	// ended up with just one
	if len(valuesToStore) == 1 {
		return valuesToStore[0]
	}

	// mix of strings and blocks
	return valuesToStore
}

func fixSingleValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		return v
	case block:
		// just in case
		if v == nil {
			return nil
		}
		return value
	default:
		panic("somehow a non-string and non-block value got into a map or list")
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
