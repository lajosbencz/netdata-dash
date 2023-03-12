/**
 * Simple object check.
 * @param item
 * @returns {boolean}
 */
const isObject = (item) => {
    return (item && typeof item === 'object' && !Array.isArray(item));
}

/**
 * Deep merge two objects.
 * @param target
 * @param ...sources
 */
const mergeDeep = (target, ...sources) => {
    if (!sources.length) return target;
    const source = sources.shift();
    if (isObject(target) && isObject(source)) {
        for (const key in source) {
            if (isObject(source[key])) {
                if (!target[key]) Object.assign(target, { [key]: {} });
                mergeDeep(target[key], source[key]);
            } else {
                Object.assign(target, { [key]: source[key] });
            }
        }
    }
    return mergeDeep(target, ...sources);
}

String.prototype.hashCode = function () {
    var hash = 0,
        i, chr;
    if (this.length === 0) return hash;
    for (i = 0; i < this.length; i++) {
        chr = this.charCodeAt(i);
        hash = ((hash << 5) - hash) + chr;
        hash |= 0; // Convert to 32bit integer
    }
    return hash;
}

const defer = () => {
    const bag = {}
    return Object.assign(
        new Promise((resolve, reject) => Object.assign(bag, { resolve, reject })),
        bag
    )
}

// const requireTag = (el, src, integrity, appendToBody) => {
//     const id = src.toString().hashCode()
//     if (document.getElementById(id) !== null) {
//         return
//     }
//     el.setAttribute('id', id);
//     el.setAttribute('referrerpolicy', 'no-referrer')
//     el.setAttribute('crossorigin', 'anonymous');
//     if (integrity) {
//         el.setAttribute('integrity', integrity)
//     }
//     (document.getElementsByTagName(!appendToBody ? 'head' : 'body')[0]).appendChild(el)
// }

// const requireJs = (src, integrity, appendToBody) => {
//     const d = defer()
//     const el = document.createElement('script')
//     el.setAttribute('type', 'text/javascript')
//     el.setAttribute('src', src)
//     el.addEventListener('load', d.resolve)
//     requireTag(el, src, integrity, appendToBody)
//     return d
// }

// const requireStyle = (src, integrity) => {
//     const el = document.createElement('link')
//     el.setAttribute('href', src)
//     el.setAttribute('rel', 'stylesheet')
//     requireTag(el, src, integrity, false)
// }
