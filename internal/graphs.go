// This file contains thin wrappers around the graph module
// for managing graph structures in the tournament data.
package internal

import (
	"iter"

	"github.com/dominikbraun/graph"
)

var nodeId int = 0

func NextNodeId() int {
	id := nodeId
	nodeId += 1
	return id
}

type GraphNode interface {
	// A unique ID that is used as the node hash
	Id() int
}

func getNodeId[T GraphNode](node T) int {
	return node.Id()
}

type DependencyGraph[T GraphNode] struct {
	graph.Graph[int, T]
	adjancencyMap map[int]map[int]graph.Edge[int]
}

func (g *DependencyGraph[T]) AddEdge(source, target T) error {
	err := g.Graph.AddEdge(source.Id(), target.Id())
	return err
}

func (g *DependencyGraph[T]) BreadthSearchIter(start T) iter.Seq2[T, int] {
	iterator := func(yield func(v T, depth int) bool) {
		visitor := func(key, depth int) bool {
			v, _ := g.Vertex(key)
			return !yield(v, depth)
		}
		graph.BFSWithDepth(g.Graph, start.Id(), visitor)
	}
	return iterator
}

// Returns the nodes that are on the outgoing edges of the given
// source node (the dependants).
func (g *DependencyGraph[T]) GetDependants(source T) []T {
	if g.adjancencyMap == nil {
		// Since the graphs do not change after their initialization
		// the adjacency map is stored on the first call
		g.adjancencyMap, _ = g.Graph.AdjacencyMap()
	}

	outEdges := g.adjancencyMap[source.Id()]
	dependants := make([]T, 0, len(outEdges))
	for k := range outEdges {
		dependant, _ := g.Vertex(k)
		dependants = append(dependants, dependant)
	}

	return dependants
}

// A RankingGraph contains all rankings of a tournament as its
// nodes. The directed edges between the nodes model the dependencies
// between the rankings.
//
// If a ranking resolves its slots from a placement in another ranking
// it will have an incoming edge from that ranking.
//
// The graph is acyclic and forms a topological hierarchy which determines
// the order in which rankings have to be updated in order to properly
// propagate a change.
type RankingGraph struct {
	DependencyGraph[Ranking]
}

func NewRankingGraph(root Ranking) *RankingGraph {
	graph := DependencyGraph[Ranking]{
		Graph: graph.New(getNodeId[Ranking], graph.Directed()),
	}
	rankingGraph := &RankingGraph{DependencyGraph: graph}
	rankingGraph.AddVertex(root)
	return rankingGraph
}

// The EliminationGraph has all matches of an elimination
// tournament as its nodes. The edges between the nodes model
// the path that the players take towards the final like
// a conventional tournament tree.
type EliminationGraph struct {
	DependencyGraph[*Match]
}

func NewEliminationGraph() *EliminationGraph {
	graph := DependencyGraph[*Match]{
		Graph: graph.New(getNodeId[*Match], graph.Directed()),
	}
	eliminationGraph := EliminationGraph{DependencyGraph: graph}
	return &eliminationGraph
}
