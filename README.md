# tiler

`tiler` is a small Sway utility that resizes the tiled windows in the focused
workspace.

It has two modes:

- `tiler`: toggle between an equal layout and a layout where the focused node
  is larger.
- `tiler --equal`: give every node the same width.

The default Sway bindings installed by `install.sh` are:

```sway
bindsym $mod+g exec ~/.local/bin/tiler
bindsym $mod+Shift+g exec ~/.local/bin/tiler --equal
```

## Layout

For two nodes, the expanded layout is approximately:

```text
33% | 67%
```

For three or more nodes, the focused node receives 50% and the remaining
width is divided equally:

```text
3 nodes: 25% | 50% | 25%
4 nodes: 16.7% | 50% | 16.7% | 16.7%
```

Expansion state is inferred from the current Sway tree. No state file is
needed, so manual resizing and Sway restarts cannot desynchronize the toggle.

Sway redistributes sibling widths after every resize command. To compensate,
`tiler` applies the complete target layout for several short convergence
passes. This produces stable equal and expanded proportions for three or more
nodes.

## Design

The program models the relevant parts of Sway directly:

- `Tree` contains the discovered workspaces.
- `Workspace` owns layout decisions such as `TargetWidths`, `EqualWidths`,
  `IsExpanded`, and `SetWidths`.
- `Node` represents a Sway container and performs targeted resizing through
  its `con_id`.
- `Rect` contains the dimensions reported by Sway.

`getTree` executes:

```bash
swaymsg -t get_tree
```

It decodes the response into recursive `Node` values, walks the tree, and
builds the list of non-scratch workspaces. The focused workspace is identified
by the focused node it contains.

Only direct tiled children of a workspace currently participate in balancing.
Floating windows are ignored. Nested split containers are discovered while
walking the Sway tree, but balancing their internal layouts is not yet
implemented.

## Extending

New layout policies should return target percentages without executing Sway
commands. For example:

```go
func (w Workspace) SomeLayout() []float64
```

The resulting widths can then be applied through:

```go
w.SetWidths(widths)
```

This keeps layout mathematics separate from Sway command execution. New CLI
modes can select another width policy in `main`, while `Node.SetWidth` remains
the single resizing boundary.

Structural layouts, such as vertical stacks around a central window, require
creating and rearranging nested Sway containers. Those should be implemented
separately from width balancing.

## Build And Test

Requirements:

- Go 1.24 or newer
- Sway and `swaymsg`

```bash
make build
go test ./...
```

The binary is written to:

```text
bin/tiler
```

Run it from an active Sway session:

```bash
make run
./bin/tiler --equal
```

## Install

Run:

```bash
./install.sh
```

The installer:

1. Checks that Go, Sway, and `swaymsg` are installed.
2. Builds and installs `tiler` as `~/.local/bin/tiler`.
3. Adds the two keybindings to the Sway config.
4. Avoids adding bindings that are already configured.
5. Refuses to overwrite conflicting `$mod+g` or `$mod+Shift+g` bindings.
6. Reloads Sway when run from an active Sway session.

If a conflicting binding exists, remove or change it explicitly and rerun the
installer. The installer does not modify unrelated bindings automatically.
