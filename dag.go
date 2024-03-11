package merkledag

import (
	"crypto/sha256"
	"encoding/json"
	"hash"
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

const MaxBlobSize = 256 * 1024 // 256KB

func Add(store KVStore, node Node, h hash.Hash) []byte {
	// 根据节点类型进行处理
	switch node.Type() {
	case FILE:
		file := node.(File)
		data := file.Bytes()

		// 如果文件大小超过256KB，则进行分割
		if len(data) > MaxBlobSize {
			var childHashes [][]byte
			for len(data) > 0 {
				chunk := data[:MaxBlobSize]
				data = data[MaxBlobSize:]

				// 计算文件块的哈希值
				chunkHash := hashChunk(chunk)
				childHashes = append(childHashes, chunkHash)

				// 存储文件块到KVStore
				err := store.Put(chunkHash, chunk)
				if err != nil {
					panic(err)
				}
			}

			// 计算文件的Merkle Root哈希值
			for _, childHash := range childHashes {
				h.Write(childHash)
			}
			fileHash := h.Sum(nil)

			// 创建文件对象的链接列表
			links := make([]Link, 0, len(childHashes))
			for _, childHash := range childHashes {
				links = append(links, Link{Hash: childHash})
			}

			// 将文件对象（包含链接列表）存储在KVStore中
			fileObj := Object{Links: links}
			objectBytes, err := json.Marshal(fileObj)
			if err != nil {
				// 处理序列化错误
				panic(err)
			}
			err = store.Put(fileHash, objectBytes)
			if err != nil {
				panic(err)
			}

			return fileHash
		}

		// 如果文件大小不超过256KB，则直接处理
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
			childHash := Add(store, child, sha256.New()) // 使用新的hash实例来计算子节点的哈希
			childHashes = append(childHashes, childHash)
		}

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

// hashChunk 计算给定数据块的哈希值
func hashChunk(chunk []byte) []byte {
	h := sha256.New()
	h.Write(chunk)
	return h.Sum(nil)
}
