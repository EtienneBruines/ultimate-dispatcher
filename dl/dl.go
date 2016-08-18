// Package dl contains the dispatcher game logic
package dl

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"engo.io/engo"
	"github.com/luxengine/math"
	yaml "gopkg.in/yaml.v2"
)

type Map struct {
	Name     string
	Nodes    []*RouteNode
	nodesMap map[uint32]*RouteNode
}

func (m *Map) Initialize() {
	m.nodesMap = make(map[uint32]*RouteNode)
	for _, node := range m.Nodes {
		m.nodesMap[node.ID] = node
	}
}

func (m *Map) AddNode(n *RouteNode) {
	m.Nodes = append(m.Nodes, n)
	m.nodesMap[n.ID] = n
}

var counter uint32 = 50000

func NewMapID() uint32 {
	counter++
	return counter
}

func RandomMap(w, h uint32, width, height float32) *Map {
	m := new(Map)
	m.Name = "RandomMap"
	m.Nodes = make([]*RouteNode, w*h)
	for i := uint32(0); i < w; i++ {
		for j := uint32(0); j < h; j++ {
			rn := new(RouteNode)
			rn.ID = i*h + (j + 1)
			rn.Location = engo.Point{
				X: float32(i+1) * width,
				Y: float32(j+1) * height,
			}
			if j > 0 {
				// Connect to every node on the left
				rn.ConnectedTo = append(rn.ConnectedTo, rn.ID-1)
			}
			if j+1 < h {
				rn.ConnectedTo = append(rn.ConnectedTo, rn.ID+1)
			}
			if i > 0 {
				// Connect to every node on the top
				rn.ConnectedTo = append(rn.ConnectedTo, rn.ID-h)
			}
			if i+1 < w {
				rn.ConnectedTo = append(rn.ConnectedTo, rn.ID+h)
			}
			m.Nodes[i*h+j] = rn
		}
	}
	return m
}

func (m *Map) Node(id uint32) *RouteNode {
	n, ok := m.nodesMap[id]
	if !ok {
		return nil
	}
	return n
}

func (m *Map) NearestNode(origin engo.Point) *RouteNode {
	var maxDistance float32 = math.MaxFloat32
	var nearestNode uint32
	for _, node := range m.Nodes {
		if d := node.Location.PointDistanceSquared(origin); d < maxDistance {
			maxDistance = d
			nearestNode = node.ID
		}
	}
	return m.Node(nearestNode)
}

func (m Map) URL() string {
	return m.Name
}

func (m Map) String() string {
	buf := bytes.NewBufferString(m.Name)
	for _, node := range m.Nodes {
		buf.WriteRune('\n')
		buf.WriteString(node.String())
	}
	return buf.String()
}

type RouteNode struct {
	ID             uint32
	Location       engo.Point
	Temporary      bool
	TemporaryUsers uint8

	ConnectedTo []uint32 `yaml:"connectedTo"`
}

func (rn RouteNode) String() string {
	return fmt.Sprintf("Node %d", rn.ID)
}

type Route struct {
	Nodes []*RouteNode
}

func (r Route) String() string {
	buf := &bytes.Buffer{}
	for _, node := range r.Nodes {
		buf.WriteString(node.String())
		buf.WriteRune('\n')
	}
	return buf.String()
}

type MapLoader struct {
	maps map[string]*Map
}

func (ml *MapLoader) Load(url string, data io.Reader) error {
	if ml.maps == nil {
		ml.maps = make(map[string]*Map)
	}

	b, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	mapDefinition := new(Map)
	yaml.Unmarshal(b, mapDefinition)
	mapDefinition.Name = url
	ml.maps[url] = mapDefinition

	return nil
}

func (ml *MapLoader) Unload(url string) error {
	delete(ml.maps, url)
	return nil
}

func (ml *MapLoader) Resource(url string) (engo.Resource, error) {
	m, exists := ml.maps[url]
	if !exists {
		return nil, fmt.Errorf("map resource was not found in memory: %s", url)
	}

	return m, nil
}
