class AStarNode {
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
