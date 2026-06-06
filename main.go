package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
)

type Tree struct {
	Workspaces []Workspace
}

func (t Tree) GetFocus() (Workspace, error) {
	for _, w := range t.Workspaces {
		if w.IsFocused() == true {
			return w, nil
		}
	}
	return Workspace{}, errors.New("No focused workspace")
}

type Rect struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Workspace struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Rect  Rect   `json:"rect"`
	Nodes []Node `json:"nodes"`
}

func (w Workspace) IsFocused() bool {
	for _, node := range w.Nodes {
		if node.Focused == true {
			return true
		}
	}
	return false
}

func (w Workspace) Print() {
	fmt.Printf("Workspace %s\n", w.Name)
	fmt.Printf("  ID:      %d\n", w.ID)
	fmt.Printf("  Rect: w=%d, h=%d\n", w.Rect.Width, w.Rect.Height)
	for _, node := range w.Nodes {
		node.Print()
	}
}

func (w Workspace) TargetWidths() []float64 {
	count := len(w.Nodes)
	widths := make([]float64, count)
	if count == 0 {
		return widths
	}

	focusWeight := math.Max(2, float64(count-1))
	if w.IsExpanded() {
		focusWeight = 1
	}
	totalWeight := float64(count-1) + focusWeight

	for i, node := range w.Nodes {
		weight := 1.0
		if node.Focused {
			weight = focusWeight
		}
		widths[i] = 100 * weight / totalWeight
	}

	return widths
}

func (w Workspace) EqualWidths() []float64 {
	widths := make([]float64, len(w.Nodes))
	for i := range widths {
		widths[i] = 100 / float64(len(widths))
	}
	return widths
}

func (w Workspace) IsExpanded() bool {
	n := len(w.Nodes)
	if n < 2 {
		return false
	}

	equalWidth := 1 / float64(n)

	for _, node := range w.Nodes {
		if node.Percent > equalWidth+0.02 {
			return true
		}
	}

	return false
}

func (w Workspace) SetWidths(widths []float64) error {
	for i, width := range widths {
		fmt.Printf("%s: %.1f%%\n", w.Nodes[i].Name, width)
	}

	for range 5 {
		for i, width := range widths {
			if err := w.Nodes[i].SetWidth(width); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w Workspace) ToggleBalance() error {
	fmt.Printf("Balancing Workspace %s\n", w.Name)
	return w.SetWidths(w.TargetWidths())
}

type Node struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Rect          Rect    `json:"rect"`
	Focused       bool    `json:"focused"`
	Percent       float64 `json:"percent"`
	Nodes         []Node  `json:"nodes"`
	FloatingNodes []Node  `json:"floating_nodes"`
}

func (n Node) Print() {
	fmt.Printf("  Node %s\n", n.Name)
	fmt.Printf("    ID:      %d\n", n.ID)
	fmt.Printf("  Rect: w=%d, h=%d\n", n.Rect.Width, n.Rect.Height)
	fmt.Printf("    Type:    %s\n", n.Type)
	fmt.Printf("    Focused: %t\n", n.Focused)
}

func (n Node) SetWidth(width float64) error {
	command := fmt.Sprintf(
		"[con_id=%d] resize set width %d ppt",
		n.ID,
		int(math.Round(width)),
	)
	out, err := exec.Command("swaymsg", command).CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize node %d: %s: %w", n.ID, out, err)
	}
	return nil
}

func getTree() (Tree, error) {
	out, err := exec.Command("swaymsg", "-t", "get_tree").Output()
	if err != nil {
		return Tree{}, err
	}

	var root Node
	if err := json.Unmarshal(out, &root); err != nil {
		return Tree{}, err
	}

	tree := Tree{}
	var walk func(Node)
	walk = func(node Node) {
		if node.Type == "workspace" && node.Name != "__i3_scratch" {
			tree.Workspaces = append(tree.Workspaces, Workspace{
				ID:    node.ID,
				Name:  node.Name,
				Rect:  node.Rect,
				Nodes: node.Nodes,
			})
		}
		for _, child := range append(node.Nodes, node.FloatingNodes...) {
			walk(child)
		}
	}
	walk(root)

	return tree, nil
}

func main() {
	tree, err := getTree()
	if err != nil {
		log.Fatal(err)
	}

	// for _, workspace := range tree.Workspaces {
	// 	workspace.Print()
	// 	fmt.Println("----------------------------------------")
	// }

	f, err := tree.GetFocus()
	if err != nil {
		log.Fatal(err)
	}
	var widths []float64
	if len(os.Args) == 2 && os.Args[1] == "--equal" {
		widths = f.EqualWidths()
	} else {
		widths = f.TargetWidths()
	}

	if err := f.SetWidths(widths); err != nil {
		log.Fatal(err)
	}
}
