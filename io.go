package prefixtreecore

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"io"
	"os"
)

// Save saves the prefixTree to an io.Writer,
// where dataType is either "json" or "gob".
// 保存将前缀树保存到io.Writer，
// 其中dataType是“ json”或“ gob”。
func (tree *PrefixTree) Save(out io.Writer, dataType string) error {
	switch dataType {
	case "gob", "GOB":
		dataEecoder := gob.NewEncoder(out)
		return dataEecoder.Encode(tree.prefixTree)
	case "json", "JSON":
		dataEecoder := json.NewEncoder(out)
		return dataEecoder.Encode(tree.prefixTree)
	}
	return ErrInvalidDataType
}

// SaveToFile saves the prefixTree to a file,
// where dataType is either "json" or "gob".
// SaveToFile将前缀树保存到文件中，
// 其中dataType是“ json”或“ gob”。
func (tree *PrefixTree) SaveToFile(fileName string, dataType string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	out := bufio.NewWriter(file)
	defer out.Flush()
	tree.Save(out, dataType)
	return nil
}

// Load loads the prefixTree from an io.Writer,
// where dataType is either "json" or "gob".
// 从io.Writer加载前缀树，
// 其中dataType是“ json”或“ gob”。
func (tree *PrefixTree) Load(in io.Reader, dataType string) error {
	switch dataType {
	case "gob", "GOB":
		dataDecoder := gob.NewDecoder(in)
		return dataDecoder.Decode(tree.prefixTree)
	case "json", "JSON":
		dataDecoder := json.NewDecoder(in)
		return dataDecoder.Decode(tree.prefixTree)
	}
	return ErrInvalidDataType
}

// LoadFromFile loads the prefixTree from a file,
// where dataType is either "json" or "gob".
// LoadFromFile从文件加载前缀树，
// 其中dataType是“ json”或“ gob”。
func (tree *PrefixTree) LoadFromFile(fileName string, dataType string) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
	defer file.Close()
	if err != nil {
		return err
	}
	in := bufio.NewReader(file)
	return tree.Load(in, dataType)
}
