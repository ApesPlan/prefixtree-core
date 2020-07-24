package prefix-tree-core

// ValueLimit 有符号整型int 最大值 二进制表示，首位0，其余1
const ValueLimit = int(^uint(0) >> 1)

type node struct {
	Value int
	Check int
}

func (n *node) base() int { return -(n.Value + 1) }

type ninfo struct {
	Sibling, Child byte
}

type block struct {
	Prev, Next, Num, Reject, Trial, Ehead int
}

func (b *block) init() {
	b.Num = 256
	b.Reject = 257
}

// PrefixTree 前缀树结构体
type PrefixTree struct {
	*prefixTree
}

type prefixTree struct {
	Array    []node
	Ninfos   []ninfo
	Blocks   []block
	Reject   [257]int
	BheadF   int
	BheadC   int
	BheadO   int
	Capacity int
	Size     int
	Ordered  bool
	MaxTrial int
}

// New 实例化 前缀树
func New() *PrefixTree {
	tree := prefixTree{
		Array:    make([]node, 256),
		Ninfos:   make([]ninfo, 256),
		Blocks:   make([]block, 1),
		Capacity: 256,
		Size:     256,
		Ordered:  true,
		MaxTrial: 1,
	}

	tree.Array[0] = node{-2, 0}
	for i := 1; i < 256; i++ {
		tree.Array[i] = node{-(i - 1), -(i + 1)}
	}
	tree.Array[1].Value = -255
	tree.Array[255].Check = -1

	tree.Blocks[0].Ehead = 1
	tree.Blocks[0].init()

	for i := 0; i <= 256; i++ {
		tree.Reject[i] = i + 1
	}

	return &PrefixTree{&tree}
}

// Get value by key, insert the key if not exist
func (tree *prefixTree) get(key []byte, from, pos int) *int {
	for ; pos < len(key); pos++ {
		if value := tree.Array[from].Value; value >= 0 && value != ValueLimit {
			to := tree.follow(from, 0)
			tree.Array[to].Value = value
		}
		from = tree.follow(from, key[pos])
	}
	to := from
	if tree.Array[from].Value < 0 {
		to = tree.follow(from, 0)
	}
	return &tree.Array[to].Value
}

func (tree *prefixTree) follow(from int, label byte) int {
	base := tree.Array[from].base()
	to := base ^ int(label)
	if base < 0 || tree.Array[to].Check < 0 {
		hasChild := false
		if base >= 0 {
			hasChild = (tree.Array[base^int(tree.Ninfos[from].Child)].Check == from)
		}
		to = tree.popEnode(base, label, from)
		tree.pushSibling(from, to^int(label), label, hasChild)
	} else if tree.Array[to].Check != from {
		to = tree.resolve(from, base, label)
	} else if tree.Array[to].Check == from {
	} else {
		panic("prefixTree: internal error, should not be here")
	}
	return to
}

func (tree *prefixTree) popBlock(bi int, head_in *int, last bool) {
	if last {
		*head_in = 0
	} else {
		b := &tree.Blocks[bi]
		tree.Blocks[b.Prev].Next = b.Next
		tree.Blocks[b.Next].Prev = b.Prev
		if bi == *head_in {
			*head_in = b.Next
		}
	}
}

func (tree *prefixTree) pushBlock(bi int, head_out *int, empty bool) {
	b := &tree.Blocks[bi]
	if empty {
		*head_out, b.Prev, b.Next = bi, bi, bi
	} else {
		tail_out := &tree.Blocks[*head_out].Prev
		b.Prev = *tail_out
		b.Next = *head_out
		*head_out, *tail_out, tree.Blocks[*tail_out].Next = bi, bi, bi
	}
}

func (tree *prefixTree) addBlock() int {
	if tree.Size == tree.Capacity {
		tree.Capacity *= 2

		oldArray := tree.Array
		tree.Array = make([]node, tree.Capacity)
		copy(tree.Array, oldArray)

		oldNinfo := tree.Ninfos
		tree.Ninfos = make([]ninfo, tree.Capacity)
		copy(tree .Ninfos, oldNinfo)

		oldBlock := tree.Blocks
		tree.Blocks = make([]block, tree.Capacity>>8)
		copy(tree.Blocks, oldBlock)
	}

	tree.Blocks[tree.Size>>8].init()
	tree.Blocks[tree.Size>>8].Ehead = tree.Size

	tree.Array[tree.Size] = node{-(tree.Size + 255), -(tree.Size + 1)}
	for i := tree.Size + 1; i < tree.Size+255; i++ {
		tree.Array[i] = node{-(i - 1), -(i + 1)}
	}
	tree.Array[tree.Size+255] = node{-(tree.Size + 254), -tree.Size}

	tree.pushBlock(tree.Size>>8, &tree.BheadO, tree.BheadO == 0)
	tree.Size += 256
	return tree.Size>>8 - 1
}

func (tree *prefixTree) transferBlock(bi int, head_in, head_out *int) {
	tree.popBlock(bi, head_in, bi == tree.Blocks[bi].Next)
	tree.pushBlock(bi, head_out, *head_out == 0 && tree.Blocks[bi].Num != 0)
}

func (tree *prefixTree) popEnode(base int, label byte, from int) int {
	e := base ^ int(label)
	if base < 0 {
		e = tree.findPlace()
	}
	bi := e >> 8
	n := &tree.Array[e]
	b := &tree.Blocks[bi]
	b.Num--
	if b.Num == 0 {
		if bi != 0 {
			tree.transferBlock(bi, &tree.BheadC, &tree.BheadF)
		}
	} else {
		tree.Array[-n.Value].Check = n.Check
		tree.Array[-n.Check].Value = n.Value
		if e == b.Ehead {
			b.Ehead = -n.Check
		}
		if bi != 0 && b.Num == 1 && b.Trial != tree.MaxTrial {
			tree.transferBlock(bi, &tree.BheadO, &tree.BheadC)
		}
	}
	n.Value = ValueLimit
	n.Check = from
	if base < 0 {
		tree.Array[from].Value = -(e ^ int(label)) - 1
	}
	return e
}

func (tree *prefixTree) pushEnode(e int) {
	bi := e >> 8
	b := &tree.Blocks[bi]
	b.Num++
	if b.Num == 1 {
		b.Ehead = e
		tree.Array[e] = node{-e, -e}
		if bi != 0 {
			tree.transferBlock(bi, &tree.BheadF, &tree.BheadC)
		}
	} else {
		prev := b.Ehead
		next := -tree.Array[prev].Check
		tree.Array[e] = node{-prev, -next}
		tree.Array[prev].Check = -e
		tree.Array[next].Value = -e
		if b.Num == 2 || b.Trial == tree.MaxTrial {
			if bi != 0 {
				tree.transferBlock(bi, &tree.BheadC, &tree.BheadO)
			}
		}
		b.Trial = 0
	}
	if b.Reject < tree.Reject[b.Num] {
		b.Reject = tree.Reject[b.Num]
	}
	tree.Ninfos[e] = ninfo{}
}

// hasChild: wherether the `from` node has children
func (tree *prefixTree) pushSibling(from, base int, label byte, hasChild bool) {
	c := &tree.Ninfos[from].Child
	keepOrder := *c == 0
	if tree.Ordered {
		keepOrder = label > *c
	}
	if hasChild && keepOrder {
		c = &tree.Ninfos[base^int(*c)].Sibling
		for tree.Ordered && *c != 0 && *c < label {
			c = &tree.Ninfos[base^int(*c)].Sibling
		}
	}
	tree.Ninfos[base^int(label)].Sibling = *c
	*c = label
}

func (tree *prefixTree) popSibling(from, base int, label byte) {
	c := &tree.Ninfos[from].Child
	for *c != label {
		c = &tree.Ninfos[base^int(*c)].Sibling
	}
	*c = tree.Ninfos[base^int(*c)].Sibling
}

func (tree *prefixTree) consult(base_n, base_p int, c_n, c_p byte) bool {
	c_n = tree.Ninfos[base_n^int(c_n)].Sibling
	c_p = tree.Ninfos[base_p^int(c_p)].Sibling
	for c_n != 0 && c_p != 0 {
		c_n = tree.Ninfos[base_n^int(c_n)].Sibling
		c_p = tree.Ninfos[base_p^int(c_p)].Sibling
	}
	return c_p != 0
}

func (tree *prefixTree) setChild(base int, c byte, label byte, flag bool) []byte {
	child := make([]byte, 0, 257)
	if c == 0 {
		child = append(child, c)
		c = tree.Ninfos[base^int(c)].Sibling
	}
	if tree.Ordered {
		for c != 0 && c <= label {
			child = append(child, c)
			c = tree.Ninfos[base^int(c)].Sibling
		}
	}
	if flag {
		child = append(child, label)
	}
	for c != 0 {
		child = append(child, c)
		c = tree.Ninfos[base^int(c)].Sibling
	}
	return child
}

func (tree *prefixTree) findPlace() int {
	if tree.BheadC != 0 {
		return tree.Blocks[tree.BheadC].Ehead
	}
	if tree.BheadO != 0 {
		return tree.Blocks[tree.BheadO].Ehead
	}
	return tree.addBlock() << 8
}

func (tree *prefixTree) findPlaces(child []byte) int {
	bi := tree.BheadO
	if bi != 0 {
		bz := tree.Blocks[tree.BheadO].Prev
		nc := len(child)
		for {
			b := &tree.Blocks[bi]
			if b.Num >= nc && nc < b.Reject {
				for e := b.Ehead; ; {
					base := e ^ int(child[0])
					for i := 0; tree.Array[base^int(child[i])].Check < 0; i++ {
						if i == len(child)-1 {
							b.Ehead = e
							return e
						}
					}
					e = -tree.Array[e].Check
					if e == b.Ehead {
						break
					}
				}
			}
			b.Reject = nc
			if b.Reject < tree.Reject[b.Num] {
				tree.Reject[b.Num] = b.Reject
			}
			bi_ := b.Next
			b.Trial++
			if b.Trial == tree.MaxTrial {
				tree.transferBlock(bi, &tree.BheadO, &tree.BheadC)
			}
			if bi == bz {
				break
			}
			bi = bi_
		}
	}
	return tree.addBlock() << 8
}

func (tree *prefixTree) resolve(from_n, base_n int, label_n byte) int {
	to_pn := base_n ^ int(label_n)
	from_p := tree.Array[to_pn].Check
	base_p := tree.Array[from_p].base()

	flag := tree.consult(base_n, base_p, tree.Ninfos[from_n].Child, tree.Ninfos[from_p].Child)
	var children []byte
	if flag {
		children = tree.setChild(base_n, tree.Ninfos[from_n].Child, label_n, true)
	} else {
		children = tree.setChild(base_p, tree.Ninfos[from_p].Child, 255, false)
	}
	var base int
	if len(children) == 1 {
		base = tree.findPlace()
	} else {
		base = tree.findPlaces(children)
	}
	base ^= int(children[0])
	var from int
	var base_ int
	if flag {
		from = from_n
		base_ = base_n
	} else {
		from = from_p
		base_ = base_p
	}
	if flag && children[0] == label_n {
		tree.Ninfos[from].Child = label_n
	}
	tree.Array[from].Value = -base - 1
	for i := 0; i < len(children); i++ {
		to := tree.popEnode(base, children[i], from)
		to_ := base_ ^ int(children[i])
		if i == len(children)-1 {
			tree.Ninfos[to].Sibling = 0
		} else {
			tree.Ninfos[to].Sibling = children[i+1]
		}
		if flag && to_ == to_pn { // new node has no child
			continue
		}
		n := &tree.Array[to]
		n_ := &tree.Array[to_]
		n.Value = n_.Value
		if n.Value < 0 && children[i] != 0 {
			// this node has children, fix their check
			c := tree.Ninfos[to_].Child
			tree.Ninfos[to].Child = c
			tree.Array[n.base()^int(c)].Check = to
			c = tree.Ninfos[n.base()^int(c)].Sibling
			for c != 0 {
				tree.Array[n.base()^int(c)].Check = to
				c = tree.Ninfos[n.base()^int(c)].Sibling
			}
		}
		if !flag && to_ == from_n { // parent node moved
			from_n = to
		}
		if !flag && to_ == to_pn {
			tree.pushSibling(from_n, to_pn^int(label_n), label_n, true)
			tree.Ninfos[to_].Child = 0
			n_.Value = ValueLimit
			n_.Check = from_n
		} else {
			tree.pushEnode(to_)
		}
	}
	if flag {
		return base ^ int(label_n)
	}
	return to_pn
}
