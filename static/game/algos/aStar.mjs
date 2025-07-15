import { minHeap } from './minHeap.mjs';

function hash(x, y) {
    //const round = (num) => Number(num.toFixed(decimals));
    //return `${round(a)},${round(b)}`;
    return `${x},${y}`;
}

function unhash(hash) {
    const [xStr, yStr] = hash.split(",");
    const X = Number(xStr);
    const Y = Number(yStr);
    return [X, Y];
}

function euclideanDistance(X1, Y1, X2, Y2) {
    const dx = X2 - X1;
    const dy = Y2 - Y1;
    return Math.sqrt(dx * dx + dy * dy);
}

function lineIntersectsBezier(X1, Y1, X2, Y2, B) {
    const line = {p1: {x: X1, y: Y1}, p2: {x: X2, y: Y2}};
    const intersects = B.intersects(line);
    if (intersects.length > 0) {
        return true
    }
    return false
}

export class aStar {

    constructor(intersections) {
        this.Nodes = new Map();
        this.Edges = new Map();
        this.Parent = new Map();
        this.openSet = new minHeap();
        this.Visited = new Map();

        for(let i = 0; i < intersections.length; i++) {

            const key = hash(intersections[i].Position[0], intersections[i].Position[1]);

            this.Nodes.set(key, Number.POSITIVE_INFINITY);

            let edgeMap = new Map();
            for(let j = 0; j < intersections[i].Edges.length; j++) {
                const key2 = hash( intersections[intersections[i].Edges[j]].Position[0], intersections[intersections[i].Edges[j]].Position[1] );
                edgeMap.set(key2, intersections[i].EdgeLengths[j])
            }
            this.Edges.set(key, edgeMap);
            this.Parent.set(key, null);
        }
    }

    addNode(X,Y) {
        const newKey = hash(X,Y);
        this.Nodes.set(newKey, Number.POSITIVE_INFINITY);

        let edgeMap = new Map();
        for(const key of this.Edges.keys()) {
            const unhashed = unhash(key);
            let valid = true;
            for(let j = 0; j < obstacles.length; j++) {
                if(lineIntersectsBezier(X, Y, unhashed[0], unhashed[1], obstacles[j])) {
                    valid = false;
                    break;
                }
            }
            if(valid) {
                const euclid = euclideanDistance(X, Y, unhashed[0], unhashed[1]);
                edgeMap.set(key, euclid);
                this.Edges.get(key).set(newKey, euclid)
            }
        }
        this.Edges.set(newKey, edgeMap);

        this.Parent.set(newKey, null);
    };

    removeNode(X, Y) {
        const newKey = hash(X,Y);
        this.Nodes.delete(newKey);

        for(const key of this.Edges.get(newKey).keys()) {
            this.Edges.get(key).delete(newKey);
        }
        this.Edges.delete(newKey);

        this.Parent.delete(newKey);
    };

    clear() {
        for (const key of this.Nodes.keys()) {
            this.Nodes.set(key, Number.POSITIVE_INFINITY);
        }

        for (const key of this.Parent.keys()) {
            this.Parent.set(key, null);
        }

        this.openSet = new minHeap();
        this.Visited = new Map();
    };

    computePath(X1, Y1, X2, Y2) {

        // if start or end do not exist, return empty
        const start = hash(X1, Y1);
        const  end = hash(X2, Y2);
        if(this.Nodes.has(start) === false || this.Nodes.has(end) === false) {
            console.log("the start or end node does not exist")
            return [];
        }

        // if start has 0 edges or end has 0 edges, no solution exists
        if(this.Edges.get(start).size === 0 || this.Edges.get(end).size === 0 ) {
            return [];
        }

        // set start to 0
        this.Nodes.set(start ,0);
        this.openSet.insert({priority: 0, key: start});

        while(this.openSet.isEmpty() === false) {

            const current = this.openSet.extractMin()

            if(current.key === end) {
                let ans = [];
                let iterate = current.key;
                while(iterate !== null) {
                    const point = unhash(iterate);
                    ans.push([point[0], point[1]]);
                    iterate = this.Parent.get(iterate);
                }

                return ans.reverse();
            }

            this.Visited.set(current.key, true);

            for(const [key, value] of this.Edges.get(current.key)) {

                if(this.Visited.has(key) === false) {

                    const unhashed = unhash(key);
                    const gVal = this.Nodes.get(current.key) + value;
                    const hVal = euclideanDistance(unhashed[0], unhashed[1], X2, Y2);
                    const fVal = gVal + hVal;

                    if(gVal < this.Nodes.get(key)) {
                        this.Parent.set(key, current.key)
                        this.Nodes.set(key, gVal);

                        if( this.openSet.has(key) ) {
                            this.openSet.update(key, fVal);
                        } else {
                            this.openSet.insert({priority: fVal, key: key})
                        }

                    }

                }

            }

        }



    };



}