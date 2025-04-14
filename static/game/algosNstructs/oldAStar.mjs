export class AStarNode {
    constructor(Origin, x, y, g = 0, h = 0, parent = null) {

        this.X = x;
        this.Y = y;
        this.G = g; // cost from start
        this.H = h; // Heuristic cost
        this.F = g + h; // Total cost
        this.parent = parent;
        this.origin = Origin;
    }
}

export function oldAStar(s, g) {
    const openList = []; // nodes to explore
    const closedList = new Set(); // nodes completely explored, prevents backtracking

    // create start and goal nodes
    let startNode = new AStarNode(s, s.X, s.Y, 0, 0, null);
    let goalNode = new AStarNode(g, g.X, g.Y, 0, 0, null);

    startNode.H = heuristic(startNode, goalNode);
    startNode.F = startNode.G + startNode.H;
    openList.push(startNode);

    while (openList.length > 0) {
        openList.sort((a, b) => a.F - b.F); // sort by lowest f value (priority queue)
        const currentNode = openList.shift();

        if ((currentNode.X == goalNode.X) && (currentNode.Y == currentNode.Y)) { // found the goal, reconstruct and return the path
            const path = [];
            let node = currentNode;
            while (node !== null) {
                path.push([node.X, node.Y]);
                node = node.parent;
            }
            return path.reverse();
        }

        closedList.add(`${currentNode.X},${currentNode.Y}`);

        for (let i = 0; i < currentNode.origin.Edges.length; i++) {

            let neighbor = null;
            console.log("test123", currentNode.origin.Edges[i])
            const firstChar = currentNode.origin.Edges[i][0];

            if (firstChar == 'd') {
                neighbor = dynamics[parseInt(currentNode.origin.Edges[i].slice(1), 10)];
            } else if (firstChar == 'i') {
                neighbor = intersections[parseInt(currentNode.origin.Edges[i].slice(1), 10)];
            }


            if (closedList.has(`${neighbor.X},${neighbor.Y}`)) {
                continue;
            }

            const tentativeG = currentNode.G + currentNode.origin.EdgeLengths[i];

            let found = null;
            for (let j = 0; j < openList.length; j++) {
                if ((openList.X == neighbor.X) && (openList.Y == neighbor.Y)) {
                    found = openList[j];
                    break;
                }
            }

            if (found == null) {
                let newNode = new AStarNode(neighbor, neighbor.X, neighbor.Y, tentativeG, heuristic(neighbor, goalNode), currentNode);
                openList.push(newNode);
            } else if (tentativeG < found.G) {
                found.G = tentativeG;
                found.H = heuristic(found, goalNode);
                found.F = found.G + found.H;
                found.parent = currentNode;
            }

        }
    }

    return [];
}
