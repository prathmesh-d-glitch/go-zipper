package huffman

type Node struct {
	Symbol    byte
	Frequency int
	Left      *Node
	Right     *Node
}

func (n *Node) IsLeaf() bool {
	return n.Left == nil && n.Right == nil
}

type minHeap struct {
	nodes []*Node
}

func newMinHeap(capacity int) *minHeap {
	return &minHeap{nodes: make([]*Node, 0, capacity)}
}

func (h *minHeap) len() int {
	return len(h.nodes)
}

func (h *minHeap) push(n *Node) {
	h.nodes = append(h.nodes, n)
	h.siftUp(len(h.nodes) - 1)
}

func (h *minHeap) pop() *Node {
	root := h.nodes[0]
	last := len(h.nodes) - 1
	h.nodes[0] = h.nodes[last]
	h.nodes[last] = nil
	h.nodes = h.nodes[:last]
	if len(h.nodes) > 0 {
		h.siftDown(0)
	}
	return root
}

func (h *minHeap) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.nodes[parent].Frequency <= h.nodes[i].Frequency {
			break
		}
		h.nodes[parent], h.nodes[i] = h.nodes[i], h.nodes[parent]
		i = parent
	}
}

func (h *minHeap) siftDown(i int) {
	n := len(h.nodes)
	for {
		smallest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && h.nodes[left].Frequency < h.nodes[smallest].Frequency {
			smallest = left
		}
		if right < n && h.nodes[right].Frequency < h.nodes[smallest].Frequency {
			smallest = right
		}
		if smallest == i {
			break
		}
		h.nodes[i], h.nodes[smallest] = h.nodes[smallest], h.nodes[i]
		i = smallest
	}
}
