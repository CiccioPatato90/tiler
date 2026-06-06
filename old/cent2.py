#!/usr/bin/env python3

import json
import logging
import subprocess
import sys

# get the focused window
# swaymsg -t get_tree | jq '.. | select(.type?) | select(.focused==true)'

# Setup logging to a file
logging.basicConfig(
    filename="/home/francesco/.config/sway/logs.log",
    level=logging.DEBUG,
    format="%(asctime)s - %(message)s",
)


def get_tree():
    output = subprocess.check_output(["swaymsg", "-t", "get_tree"])
    return json.loads(output)


def get_current_workspace():
    output = subprocess.check_output(
        ["swaymsg", "-t", "get_outputs", "| jq '.[] | .current_workspace'"]
    )
    print(output)
    return json.loads(output)


def find_focused_and_workspace(node, workspace=None):
    if node.get("type") == "workspace":
        workspace = node

    if node.get("focused"):
        return node, workspace

    # Recursively search children
    for child in node.get("nodes", []) + node.get("floating_nodes", []):
        result = find_focused_and_workspace(child, workspace)
        if result:
            return result
    return None


def main():
    tree = get_tree()
    focused, workspace = find_focused_and_workspace(tree)

    if not focused or not workspace:
        sys.exit(1)

    # 1. Identify Tiled Windows
    # We filter out floating windows to get the true "tiled" layout indices
    tiled_nodes = [n for n in workspace["nodes"] if n["type"] == "con"]
    count = len(tiled_nodes)
    logging.debug("[PRESTART] %s.", tiled_nodes)

    if count <= 1:
        sys.exit(0)

    # 2. Find Current State
    # Find the index (0, 1, 2) of the focused window
    try:
        # We search by ID to find exactly which node in the list is 'focused'
        current_index = next(
            i for i, v in enumerate(tiled_nodes) if v["id"] == focused["id"]
        )
        logging.debug(f"[current_index]. {current_index}")
    except StopIteration:
        logging.debug("[Stopping iteration].")
        sys.exit(1)

    # Calculate width percentage
    parent_width = workspace["rect"]["width"]
    focused_width = focused["rect"]["width"]
    current_pct = (focused_width / parent_width) * 100

    cmds = []

    # --- LOGIC ---
    logging.debug("[START].")
    # CASE: 2 Windows
    if count == 2:
        logging.debug("[INFO] 2 Windows detected.")
        # Target: Right side (Index 1)
        # Target Size: 2/3 (66%)

        # Check if we are already expanded (Toggle logic)
        if 65 < current_pct < 68:
            # We are big -> Reset to equal (50%)
            cmds.append("resize set width 50ppt")
            logging.debug("[INFO] Toggled. Restored equal proportions.")
        else:
            # We are small -> Move to Right & Resize to 66%
            if current_index != 1:
                # Swap with the window currently at index 1
                target_id = tiled_nodes[1]["id"]
                cmds.append(f"swap container with con_id {target_id}")
                logging.debug("[INFO] Moving focus to center.")

            cmds.append("resize set width 66 ppt")
            logging.debug("[INFO] Setting center to large.")

    # CASE: 3 Windows (or more)
    elif count >= 3:
        # Target: Center (Index 1)
        # Target Size: 1/2 (50%)
        logging.debug("[INFO] 3 Windows detected.")
        # Check if we are already expanded (Toggle logic)
        if 49 < current_pct < 51:
            # We are big -> Reset to equal (33%)
            logging.debug("[INFO] Toggle. Restored equal proportions.")
            cmds.append("resize set width 33 ppt")
        else:
            # We are small -> Move to Center & Resize to 50%
            if current_index != 1:
                # Swap with the window currently at index 1 (the middle one)
                target_id = tiled_nodes[1]["id"]
                cmds.append(f"swap container with con_id {target_id}")
                logging.debug("[INFO] Moving Focus to center.")

            cmds.append("resize set width 50 ppt")
            logging.debug("[INFO] Setting center to large.")

    # 3. Execute Commands
    # We join commands with ";" to execute them in one go
    if cmds:
        for cmd in cmds:
            logging.debug(f"[INFO] Executing: {cmd}.")
        full_cmd = "; ".join(cmds)

        subprocess.run(["swaymsg", full_cmd])

        logging.debug("[END].")


if __name__ == "__main__":
    main()
