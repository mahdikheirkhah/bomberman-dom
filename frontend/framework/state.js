/**
 * Simple state management system
 */
class Store {
    /**
     * Creates a new store instance
     * @param {Object} initialState - Initial state object
     */
    constructor(initialState = {}) {
        this.state = initialState;
        this.listeners = new Set();
    }

    /**
     * Gets the current state
     * @returns {Object} Current state
     */
    getState() {
        return this.state;
    }

    /**
     * Updates the state
     * @param {Object|Function} update - New state object or update function
     */
    setState(update) {
        const newState = typeof update === 'function'
            ? update(this.state)
            : update;
        this.state = { ...this.state, ...newState };
        this.notify();
    }

    /**
    * Removes a property from the state
    * @param {string} key - Property name to remove
    */
    removeState(key) {
        const { [key]: _, ...rest } = this.state;
        this.state = rest;
        this.notify();
    }

    /**
     * Subscribe to state changes
     * @param {Function} listener - Callback function
     * @returns {Function} Unsubscribe function
     */
    subscribe(listener) {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener);
    }

    /**
     * Subscribe to state changes
     * @param {Function} listener - Callback function
     */
    unsubscribe(listener) {
        this.listeners.delete(listener);
    }

    /**
     * Notify all listeners of state change
     */
    notify() {
        for (const listener of this.listeners) {
            listener(this.state);
        }
    }
}

// Create a global store instance
const store = new Store();

export { Store, store };