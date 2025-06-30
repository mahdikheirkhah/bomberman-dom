// Import and re-export all framework components
import { createElement, mount } from './dom.js';
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
    

    render = () => {
        const newVnode = this.rootComponent();
        this.container.innerHTML = '';
        if (Array.isArray(newVnode)) {
            newVnode.forEach(vnode => mount(vnode, this.container));
        } else {
            mount(newVnode, this.container);
        }
        this.vnode = newVnode;
    }
    
    start() {
        this.unsubscribe = store.subscribe(this.render);
        this.render();
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
};