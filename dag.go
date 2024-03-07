package merkledag

import (
	"encoding/json"
	"hash"
	"sort"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

func Add(store KVStore, node Node, h hash.Hash) []byte {
	// 根据节点类型进行处理
	switch node.Type() {
	case FILE:
		file := node.(File)
		data := file.Bytes()

		// 计算文件内容的哈希值
		h.Write(data)
		fileHash := h.Sum(nil)

		// 将文件对象存储在KVStore中
		fileObj := Object{Data: data}
		err := store.Put(fileHash, fileObj.Data)
		if err != nil {
			panic(err)
		}

		return fileHash

	case DIR:
		dir := node.(Dir)
		it := dir.It()

		var childHashes [][]byte
		for it.Next() {
			child := it.Node()
			var newH hash.Hash
			childHash := Add(store, child, newH) // 使用h的新实例来计算子节点的哈希
			childHashes = append(childHashes, childHash)
		}

		// 对子节点的哈希值进行排序
		sortHashes(childHashes)

		// 计算目录的Merkle Root哈希值
		for _, childHash := range childHashes {
			h.Write(childHash)
		}
		dirHash := h.Sum(nil)

		// 将目录对象存储在KVStore中
		links := make([]Link, 0, len(childHashes))
		for _, childHash := range childHashes {
			links = append(links, Link{Hash: childHash})
		}
		dirObj := Object{Links: links}
		objectBytes, err := json.Marshal(dirObj)
		if err != nil {
			// 处理序列化错误
			panic(err)
		}
		err = store.Put(dirHash, objectBytes)
		if err != nil {
			panic(err)
		}

		return dirHash

	default:
		panic("unsupported node type")
	}
}

// sortHashes 是一个辅助函数，用于对哈希值进行排序
func sortHashes(hashes [][]byte) {
	sort.Slice(hashes, func(i, j int) bool {
		return compare(hashes[i], hashes[j]) < 0
	})
}

// compare 是一个辅助函数，用于比较两个哈希值
func compare(a, b []byte) int {
	for i := range a {
		if a[i] != b[i] {
			return int(a[i]) - int(b[i])
		}
	}
	return len(a) - len(b)
}
