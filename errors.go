package prefix-tree-core

import "errors"

var (
	ErrInvalidDataType = errors.New("prefixTree: invalid datatype")
	ErrInvalidValue    = errors.New("prefixTree: invalid value")
	ErrInvalidKey      = errors.New("prefixTree: invalid key")
	ErrNoPath          = errors.New("prefixTree: no path")
	ErrNoValue         = errors.New("prefixTree: no value")
)
