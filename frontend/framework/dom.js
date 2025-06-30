
/**
 * Creates a virtual DOM node
 * @param {string} tag - HTML tag name
 * @param {Object} attrs - Element attributes
 * @param {Array} children - Child elements
 * @returns {Object} Virtual DOM node
 */
function createElement(tag, attrs = {}, children = []) {
    return {
        tag,
        attrs,
        children: children.flat().map(child =>
            typeof child === 'string' || typeof child === 'number'
                ? { tag: 'text', attrs: {}, text: String(child) }
                : child
        )
    };
}

/**
 * Applies attributes to a DOM element
 * @param {HTMLElement} element - The DOM element
 * @param {Object} attrs - Attributes to apply
 */
function applyAttributes(element, attrs) {
    if (!element || !(element instanceof Element)) {
        console.error('Invalid DOM element for attributes:', element);
        return;
    }

    if (!attrs || typeof attrs !== 'object') {
        console.warn('Invalid attributes object:', attrs);
        return;
    }

    try {
        for (const [key, value] of Object.entries(attrs)) {
            if (key.startsWith('on')) {
                element[key] = value;
            } else if (key === 'ref') {
                // Handle refs
                if (typeof value === 'function') {
                    value(element);
                }
            } else if (key === 'className') {
                // Special case for class
                element.setAttribute('class', value);
            } else if (typeof value === 'boolean') {
                // Handle boolean attributes
                if (value) {
                    element.setAttribute(key, '');
                } else {
                    element.removeAttribute(key);
                }
            } else if (value !== null && value !== undefined) {
                // Handle regular attributes
                element.setAttribute(key, String(value));
            }
        }
    } catch (error) {
        console.error('Error applying attributes:', error);
    }
}

/**
 * Creates a real DOM element from a virtual node
 * @param {Object} vnode - Virtual DOM node
 * @returns {HTMLElement} Real DOM element
 */
function createDomElement(vnode) {
    // Handle null, undefined, or invalid vnodes
    if (!vnode || typeof vnode !== 'object') {
        return document.createTextNode('');
    }

    // Ensure vnode has required properties
    if (!('tag' in vnode)) {
        console.error('Missing tag in vnode:', vnode);
        return document.createTextNode('');
    }

    if (vnode.tag === 'text') {
        return document.createTextNode(vnode.text || '');
    }

    if (typeof vnode.tag !== 'string') {
        console.error('Invalid vnode tag type:', typeof vnode.tag);
        return document.createTextNode('');
    }

    try {
        const element = document.createElement(vnode.tag);
        applyAttributes(element, vnode.attrs || {});

        const children = Array.isArray(vnode.children) ? vnode.children.filter(child => child !== null && child !== undefined) : [];
        for (const child of children) {
            mount(child, element);
        }

        return element;
    } catch (error) {
        console.error('Error creating DOM element:', error);
        return document.createTextNode('');
    }
}

/**
 * Mounts a virtual node to a parent DOM element
 * Creates DOM elements recursively from virtual DOM tree
 * @param {Object} vnode - Virtual DOM node
 * @param {HTMLElement} parentDomElement - Parent DOM element
 * @returns {HTMLElement} Created DOM element
 */
function mount(vnode, parentDomElement) {
    const domElement = createDomElement(vnode);
    if (!parentDomElement) {
        console.error('Mount failed: parentDomElement is null/undefined');
        return domElement;
    }
    parentDomElement.appendChild(domElement);
    return domElement;
}



// Export the functions
export { createElement, mount };