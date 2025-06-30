/**
 * Simple routing system
 */
class Router {
    constructor() {
        this.routes = new Map();
        this.currentRoute = null;
        this.defaultHandler = () => {
            console.warn('No default route handler set');
        };

        // Listen for hash changes
        window.addEventListener('hashchange', () => {
            this.handleRoute(this.getCurrentPath());
        });
    }
        
    getCurrentPath() {
        // Remove the leading '#' from the hash
        const hash = window.location.hash || '#/';
        return hash.slice(1); // removes '#'
    }
    /**
     * Add a route
     * @param {string} path - URL path
     * @param {Function} handler - Route handler function
     */
    addRoute(path, handler) {
        if (typeof handler !== 'function') {
            throw new Error('Route handler must be a function');
        }
        this.routes.set(path, handler);
    }

    /**
     * Set default route handler
     * @param {Function} handler - Default route handler function
     */
    setDefaultHandler(handler) {
        if (typeof handler !== 'function') {
            throw new Error('Default handler must be a function');
        }
        this.defaultHandler = handler;
    }

    /**
     * Navigate to a route
     * @param {string} path - URL path
     */
    navigate(path) {
        // This triggers the hashchange event
        window.location.hash = path.startsWith('/') ? `#${path}` : `#/${path}`;
    }

    /**
     * Handle route change
     * @param {string} path - URL path
     */
    handleRoute(path) {
        const handler = this.routes.get(path);
        if (handler) {
            this.currentRoute = path;
            handler();
        } else {
            this.currentRoute = '/';
            this.defaultHandler();
        }
    }


    /**
     * Initialize router with current path
     */
    init() {
        if (!this.routes.has('/')) {
            this.addRoute('/', this.defaultHandler);
        }
        this.handleRoute(this.getCurrentPath());
    }
}

// Create a global router instance
const router = new Router();

export { Router, router };