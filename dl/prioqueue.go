package dl

import "sort"

type PriorityQueue struct {
	list   []interface{}
	values []float32
}

func (p *PriorityQueue) Enqueue(item interface{}, value float32) {
	index := SearchFloat32s(p.values, value)
	p.values = append(p.values[:index], append([]float32{value}, p.values[index:]...)...)
	p.list = append(p.list[:index], append([]interface{}{item}, p.list[index:]...)...)
}

func (p *PriorityQueue) Dequeue() interface{} {
	p.values = p.values[1:]
	item := p.list[0]
	p.list = p.list[1:]
	return item
}

func SearchFloat32s(a []float32, x float32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}
