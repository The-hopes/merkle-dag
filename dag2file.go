package merkledag

import (
	"encoding/json"
)

// Hash to file
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
	data, err := store.Get(hash)
	if err != nil {
		panic(err)
	}
	var tree Object
	if err := json.Unmarshal(data, &tree); err != nil {
		//反序列化失败
		panic(err)
	}

	// 遍历Links找到匹配的path
	for _, link := range tree.Links {
		if link.Name == path {
			// 根据Link的的hash从KVStore中检索文件内容
			fileContent, err := store.Get(link.Hash)
			if err != nil {
				panic(err) // 表示检索文件内容失败
			}
			return fileContent // 返回文件内容
		}
	}
	return nil
}
