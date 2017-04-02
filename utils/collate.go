package utils

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
)

// turn arrays of of strings, numbers or whatever into byte keys to be used in levelup
func ToIndexable(x interface{}) (result []byte) {
	result = append(result, collationIndex(x))
	result = append(result, indexify(x)...)
	return append(result, 0)
}

func collationIndex(x interface{}) byte {
	if x == nil {
		return '1'
	}
	switch x.(type) {
	case bool:
		return '2'
	case float64, int:
		return '3'
	case string:
		return '4'
	case []interface{}:
		return '5'
	case map[interface{}]interface{}:
		return '6'
	default:
		panic("collationIndex does not work with the type " + reflect.TypeOf(x).Name())
	}
}

func indexify(key interface{}) (result []byte) {
	switch k := key.(type) {
	case bool:
		if k {
			return []byte{'1'}
		} else {
			return []byte{'0'}
		}
	case float64:
		return numToIndexable(k)
	case int:
		return numToIndexable(float64(k))
	case string:
		return bytes.Replace(
			bytes.Replace(
				bytes.Replace(
					[]byte(k),
					[]byte{'2'}, []byte{'2', '2'}, -1),
				[]byte{'1'}, []byte{'1', '2'}, -1),
			[]byte{'0'}, []byte{'1', '1'}, -1)
	case []interface{}:
		for _, part := range k {
			result = append(result, ToIndexable(part)...)
		}
		return result
	case map[interface{}]interface{}:
		for part1, part2 := range k {
			result = append(result, ToIndexable(part1)...)
			result = append(result, ToIndexable(part2)...)
		}
		return result
	}
	panic("indexify does not work with the type " + reflect.TypeOf(key).Name())
}

func numToIndexable(num float64) (result []byte) {
	if num == 0 {
		return []byte{'1'}
	}

	// convert number to exponential format for easier and
	// more succinct string sorting
	expFormat := strings.Split(strconv.FormatFloat(num, 'e', -1, 64), "e+")
	magnitude, _ := strconv.Atoi(expFormat[1])

	neg := num < 0

	if neg {
		result = append(result, '0')
		magnitude = -magnitude
	} else {
		result = append(result, '2')
	}

	// first sort by magnitude
	// it's easier if all magnitudes are positive
	magnitudeString := strconv.Itoa(magnitude + 324)
	pad := 3 - len(magnitudeString)
	magnitudeString = strings.Repeat("0", pad) + magnitudeString

	result = append(result, []byte(magnitudeString)...)

	// then sort by the factor
	factor, _ := strconv.ParseFloat(expFormat[0], 32) // [1..10]
	if neg {
		// for negative reverse ordering
		factor = 10 - factor
	}

	factorStr := strconv.FormatFloat(factor, 'f', -1, 64)

	// strip zeros from the end
	factorStr = strings.TrimRight(factorStr, "0.")

	result = append(result, []byte(factorStr)...)

	return result
}
