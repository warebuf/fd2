export class minHeap {
    constructor() {
        this.heap = [];
        this.indexMap = new Map(); // maps keys to indices
    }

    insert(node) {
        this.heap.push(node);
        this.indexMap.set(node.key, this.heap.length - 1);
        this._bubbleUp(this.heap.length - 1);
    }

    extractMin() {
        if (this.heap.length === 0) return null;
        if (this.heap.length === 1) {
            const min = this.heap.pop();
            this.indexMap.delete(min.key);
            return min;
        }

        const min = this.heap[0];
        const end = this.heap.pop();
        this.heap[0] = end;
        this.indexMap.set(end.key, 0);
        this.indexMap.delete(min.key);
        this._bubbleDown(0);

        return min;
    }

    update(key, newPriority) {
        const index = this.indexMap.get(key);
        if (index === undefined) return;

        const node = this.heap[index];
        const oldPriority = node.priority;
        node.priority = newPriority;

        if (newPriority < oldPriority) {
            this._bubbleUp(index);
        } else {
            this._bubbleDown(index);
        }
    }

    _bubbleUp(index) {
        while (index > 0) {
            let parent = Math.floor((index - 1) / 2);
            if (this.heap[index].priority >= this.heap[parent].priority) break;

            this._swap(index, parent);
            index = parent;
        }
    }

    _bubbleDown(index) {
        const length = this.heap.length;
        while (true) {
            let left = 2 * index + 1;
            let right = 2 * index + 2;
            let smallest = index;

            if (left < length && this.heap[left].priority < this.heap[smallest].priority) {
                smallest = left;
            }
            if (right < length && this.heap[right].priority < this.heap[smallest].priority) {
                smallest = right;
            }
            if (smallest === index) break;

            this._swap(index, smallest);
            index = smallest;
        }
    }

    _swap(i, j) {
        [this.heap[i], this.heap[j]] = [this.heap[j], this.heap[i]];
        this.indexMap.set(this.heap[i].key, i);
        this.indexMap.set(this.heap[j].key, j);
    }

    isEmpty() {
        return this.heap.length === 0;
    }

    has(key) {
        return this.indexMap.has(key);
    }
}
