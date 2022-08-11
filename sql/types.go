package sql

import (
	"errors"
	"fmt"
)

var Types = map[string]int{
	"int":       11,
	"bigint":    20,
	"varchar":   255,
	"text":      0,
	"date":      0,
	"time":      0,
	"timestamp": 0,
	"float":     0,
	"json":      0,
	"enum":      0,
}

var TypeDict = map[string]string{
	"float32":         "float",
	"float64":         "float",
	"int":             "int",
	"uint":            "int",
	"int64":           "bigint",
	"uint64":          "bigint",
	"string":          "varchar",
	"time.Time":       "timestamp",
	"bool":            "tinyint",
	"[]byte":          "blob",
	"[]uint8":         "blob",
	"*float32":        "float",
	"*float64":        "float",
	"*int":            "int",
	"*uint":           "int",
	"*int64":          "bigint",
	"*uint64":         "bigint",
	"*string":         "varchar",
	"*time.Time":      "timestamp",
	"*bool":           "tinyint",
	"*[]byte":         "blob",
	"*[]uint8":        "blob",
	"sql.NullFloat32": "float",
	"sql.NullFloat64": "float",
	"sql.NullInt64":   "bigint",
	"sql.NullString":  "varchar",
	"sql.NullBool":    "tinyint",
}

func MatchType(typeName string) (string, error) {
	if result, ok := TypeDict[typeName]; ok {
		return result, nil
	}

	return "", errors.New(fmt.Sprintf("[crud] Can't match Go type '%s' with any SQL type.", typeName))
}
