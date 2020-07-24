package prefixtreecore

// Status reports the following statistics of the prefixTree:
//	keys:		number of keys that are in the prefixTree,
//	nodes:		number of trie nodes (slots in the base array) has been taken,
//	size:			the size of the base array used by the prefixTree,
//	capacity:		the capicity of the base array used by the prefixTree.
// 状态报告前缀树的以下统计信息：
// 键：前缀树中的键数，
// 节点：已获取Trie节点（基本数组中的插槽）的数量，
// size：前缀树使用的基本数组的大小，
// 容量：前缀树使用的基本数组的容量。
func (tree *PrefixTree) Status() (keys, nodes, size, capacity int) {
	for i := 0; i < tree.Size; i++ {
		n := tree.Array[i]
		if n.Check >= 0 {
			nodes++
			if n.Value >= 0 {
				keys++
			}
		}
	}
	return keys, nodes, tree.Size, tree.Capacity
}

// Jump travels from a node `from` to another node `to` by following the path `path`.
// For example, if the following keys were inserted:
// 按照路径`path`跳转从一个节点`from`到另一个节点`to`。
// 例如，如果插入了以下键：
//	id	key
//	19	abc
//	23	ab
//	37	abcd
// then
//	Jump([]byte("ab"), 0) = 23, nil		// reach "ab" from root
//	Jump([]byte("c"), 23) = 19, nil			// reach "abc" from "ab"
//	Jump([]byte("cd"), 23) = 37, nil		// reach "abcd" from "ab"
func (tree *PrefixTree) Jump(path []byte, from int) (to int, err error) {
	for _, b := range path {
		if tree.Array[from].Value >= 0 {
			return from, ErrNoPath
		}
		to = tree.Array[from].base() ^ int(b)
		if tree.Array[to].Check != from {
			return from, ErrNoPath
		}
		from = to
	}
	return to, nil
}

// Key returns the key of the node with the given `id`.
// It will return ErrNoPath, if the node does not exist.
// Key返回具有给定id的节点的Key。
// 如果该节点不存在，它将返回ErrNoPath。
func (tree *PrefixTree) Key(id int) (key []byte, err error) {
	for id > 0 {
		from := tree.Array[id].Check
		if from < 0 {
			return nil, ErrNoPath
		}
		if char := byte(tree.Array[from].base() ^ id); char != 0 {
			key = append(key, char)
		}
		id = from
	}
	if id != 0 || len(key) == 0 {
		return nil, ErrInvalidKey
	}
	for i := 0; i < len(key)/2; i++ {
		key[i], key[len(key)-i-1] = key[len(key)-i-1], key[i]
	}
	return key, nil
}

// Value returns the value of the node with the given `id`.
// It will return ErrNoValue, if the node does not have a value.
// Value返回具有给定id的节点的值。
// 如果节点没有值，它将返回ErrNoValue。
func (tree *PrefixTree) Value(id int) (value int, err error) {
	value = tree.Array[id].Value
	if value >= 0 {
		return value, nil
	}
	to := tree.Array[id].base()
	if tree.Array[to].Check == id && tree.Array[to].Value >= 0 {
		return tree.Array[to].Value, nil
	}
	return 0, ErrNoValue
}

// Insert adds a key-value pair into the prefixTree.
// It will return ErrInvalidValue, if value < 0 or >= ValueLimit.
// Insert将一个键值对添加到前缀树中。
// 如果值 < 0或 >= ValueLimit，它将返回ErrInvalidValue。
func (tree *PrefixTree) Insert(key []byte, value int) error {
	if value < 0 || value >= ValueLimit {
		return ErrInvalidValue
	}
	p := tree.get(key, 0, 0)
	*p = value
	return nil
}

// Update increases the value associated with the `key`.
// The `key` will be inserted if it is not in the prefixTree.
// It will return ErrInvalidValue, if the updated value < 0 or >= ValueLimit.
// Update会增加与`key`相关联的值。
// 如果不在前缀树中，则将插入`key`。
// 如果更新后的值<0或> = ValueLimit，它将返回ErrInvalidValue。
func (tree *PrefixTree) Update(key []byte, value int) error {
	p := tree.get(key, 0, 0)

	// key was not inserted
	if *p == ValueLimit {
		*p = value
		return nil
	}

	// key was inserted before
	if *p+value < 0 || *p+value >= ValueLimit {
		return ErrInvalidValue
	}
	*p += value
	return nil
}

// Delete removes a key-value pair from the prefixTree.
// It will return ErrNoPath, if the key has not been added.
// Delete从前缀树中删除键/值对。
// 如果尚未添加密钥，它将返回ErrNoPath。
func (tree *PrefixTree) Delete(key []byte) error {
	// if the path does not exist, or the end is not a leaf, nothing to delete
	// f路径不存在，或者末尾不是叶子，无可删除
	to, err := tree.Jump(key, 0)
	if err != nil {
		return ErrNoPath
	}

	if tree.Array[to].Value < 0 {
		base := tree.Array[to].base()
		if tree.Array[base].Check == to {
			to = base
		}
	}

	for to > 0 {
		from := tree.Array[to].Check
		base := tree.Array[from].base()
		label := byte(to ^ base)

		// if `to` has sibling, remove `to` from the sibling list, then stop
		// 如果`to`有同级，请从同级列表中删除`to`，然后停止
		if tree.Ninfos[to].Sibling != 0 || tree.Ninfos[from].Child != label {
			// delete the label from the child ring first
			// 首先从子环删除标签
			tree.popSibling(from, base, label)
			// then release the current node `to` to the empty node ring
			// 然后将当前节点`to`释放到空节点环
			tree.pushEnode(to)
			break
		}
		// otherwise, just release the current node `to` to the empty node ring
		// 否则，只需将当前节点`to`释放到空节点环
		tree.pushEnode(to)
		// then check its parent node
		// 然后检查其父节点
		to = from
	}
	return nil
}

// Get returns the value associated with the given `key`.
// It is equivalent to
// Get返回与给定`key`相关联的值。
// 相当于
//		id, err1 = Jump(key)
//		value, err2 = Value(id)
// Thus, it may return ErrNoPath or ErrNoValue,
// 因此，它可能会返回ErrNoPath或ErrNoValue，
func (tree *PrefixTree) Get(key []byte) (value int, err error) {
	to, err := tree.Jump(key, 0)
	if err != nil {
		return 0, err
	}
	return tree.Value(to)
}

// PrefixMatch returns a list of at most `num` nodes which match the prefix of the key.
// If `num` is 0, it returns all matches.
// For example, if the following keys were inserted:
// PrefixMatch返回最多与关键字的前缀相匹配的“ num”个节点的列表。
// 如果`num`为0，则返回所有匹配项。
// 例如，如果插入了以下键：
//	id	key
//	19	abc
//	23	ab
//	37	abcd
// then
//	PrefixMatch([]byte("abc"), 1) = [ 23 ]				// match ["ab"]
//	PrefixMatch([]byte("abcd"), 0) = [ 23, 19, 37]		// match ["ab", "abc", "abcd"]
func (tree *PrefixTree) PrefixMatch(key []byte, num int) (ids []int) {
	for from, i := 0, 0; i < len(key); i++ {
		to, err := tree.Jump(key[i:i+1], from)
		if err != nil {
			break
		}
		if _, err := tree.Value(to); err == nil {
			ids = append(ids, to)
			num--
			if num == 0 {
				return
			}
		}
		from = to
	}
	return
}

// PrefixPredict returns a list of at most `num` nodes which has the key as their prefix.
// These nodes are ordered by their keys.
// If `num` is 0, it returns all matches.
// For example, if the following keys were inserted:
// PrefixPredict返回一个最多有num个节点的列表，该列表以键作为前缀。
// 这些节点按其键排序。
// 如果`num`为0，则返回所有匹配项。
// 例如，如果插入了以下键：
//	id	key
//	19	abc
//	23	ab
//	37	abcd
// then
//	PrefixPredict([]byte("ab"), 2) = [ 23, 19 ]			// predict ["ab", "abc"]
//	PrefixPredict([]byte("ab"), 0) = [ 23, 19, 37 ]		// predict ["ab", "abc", "abcd"]
func (tree *PrefixTree) PrefixPredict(key []byte, num int) (ids []int) {
	root, err := tree.Jump(key, 0)
	if err != nil {
		return
	}
	for from, err := tree.begin(root); err == nil; from, err = tree.next(from, root) {
		ids = append(ids, from)
		num--
		if num == 0 {
			return
		}
	}
	return
}

func (tree *PrefixTree) begin(from int) (to int, err error) {
	for c := tree.Ninfos[from].Child; c != 0; {
		to = tree.Array[from].base() ^ int(c)
		c = tree.Ninfos[to].Child
		from = to
	}
	if tree.Array[from].base() > 0 {
		return tree.Array[from].base(), nil
	}
	return from, nil
}

func (tree *PrefixTree) next(from int, root int) (to int, err error) {
	c := tree.Ninfos[from].Sibling
	for c == 0 && from != root && tree.Array[from].Check >= 0 {
		from = tree.Array[from].Check
		c = tree.Ninfos[from].Sibling
	}
	if from == root {
		return 0, ErrNoPath
	}
	from = tree.Array[tree.Array[from].Check].base() ^ int(c)
	return tree.begin(from)
}
