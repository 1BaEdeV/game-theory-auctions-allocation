package main

import (
	"sort"
)

type Graph struct {
	Nodes map[int]bool
	Edges map[int]map[int]bool
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[int]bool),
		Edges: make(map[int]map[int]bool),
	}
}

func (g *Graph) AddNode(node int) {
	g.Nodes[node] = true
	if g.Edges[node] == nil {
		g.Edges[node] = make(map[int]bool)
	}
}

func (g *Graph) AddEdge(u, v int) {
	g.AddNode(u)
	g.AddNode(v)
	g.Edges[u][v] = true
	g.Edges[v][u] = true
}

func (g *Graph) GetNeighbors(node int) []int {
	var neighbors []int
	for nghbr := range g.Edges[node] {
		neighbors = append(neighbors, nghbr)
	}
	sort.Ints(neighbors)
	return neighbors
}

func (g *Graph) HasEdge(u, v int) bool {
	return g.Edges[u][v]
}

func (g *Graph) NumNodes() int {
	return len(g.Nodes)
}

func (g *Graph) NumEdges() int {
	count := 0
	for u := range g.Edges {
		count += len(g.Edges[u])
	}
	return count / 2
}

func (g *Graph) GetNodeList() []int {
	var nodes []int
	for node := range g.Nodes {
		nodes = append(nodes, node)
	}
	sort.Ints(nodes)
	return nodes
}

// NumCommunities считает количество уникальных сообществ
