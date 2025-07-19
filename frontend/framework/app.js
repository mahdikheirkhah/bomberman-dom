// Import and re-export all framework components
import { createElement, mount, patchChildren } from './dom.js';
import { Store, store } from './state.js';
import { Router, router } from './router.js';

// Framework version
const VERSION = '1.0.0';

// App renderer class for managing component lifecycle
class AppRenderer {
    constructor(rootComponent, container) {
        this.rootComponent = rootComponent;
        this.container = container;
        this.vnode = null;
        this.unsubscribe = null;
    }

    render = (force = false) => {
        if (!force && store.getState().preventRender) {
            store.setState({ preventRender: false }, true);
            return;
        }

        const newVnodeOrVnodes = this.rootComponent();
        const oldVnodeOrVnodes = this.vnode;

        if (oldVnodeOrVnodes) {
            const oldVnodes = Array.isArray(oldVnodeOrVnodes) ? oldVnodeOrVnodes : [oldVnodeOrVnodes];
            const newVnodes = Array.isArray(newVnodeOrVnodes) ? newVnodeOrVnodes : [newVnodeOrVnodes];
            patchChildren(this.container, oldVnodes, newVnodes);
        } else {
            const newVnodes = Array.isArray(newVnodeOrVnodes) ? newVnodeOrVnodes : [newVnodeOrVnodes];
            newVnodes.forEach(vnode => mount(vnode, this.container));
        }

        this.vnode = newVnodeOrVnodes;
    }

    start() {
        this.unsubscribe = store.subscribe(this.render);
        this.render(true);
    }

    destroy() {
        if (this.unsubscribe) {
            this.unsubscribe();
            this.unsubscribe = null;
        }
    }
}

// Helper function to create and mount an app
function createApp(rootComponent, container) {
    const renderer = new AppRenderer(rootComponent, container);
    
    renderer.start();
    router.init();
    
    return {
        store,
        router,
        destroy: () => renderer.destroy()
    };
}

export {
    createElement,
    createApp,
    Store,
    store,
    Router,
    router,
    VERSION
}