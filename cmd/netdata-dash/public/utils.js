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

const isObject = (item) => {
    return (item && typeof item === 'object' && !Array.isArray(item));
}

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

const defer = () => {
    const bag = {}
    return Object.assign(
        new Promise((resolve, reject) => Object.assign(bag, { resolve, reject })),
        bag
    )
}

const observeNodeRemoved = (node, callback) => {
    const o = new MutationObserver(function (ms) {
        ms.forEach(function (m) {
            m.removedNodes.forEach(function (n) {
                if (n === node) {
                    try {
                        callback()
                    }
                    catch (e) {
                        console.error(e)
                    }
                    o.disconnect()
                }
            })
        })
    })
    o.observe(node.parentNode, { subtree: false, childList: true })
}

const formatDate = (dt) => {
    return dt.toISOString().substring(0, 19)
}

const smartFormatDate = (dt) => {
    const now = new Date()
    let str = ""
    const pad = v => v.toString().padStart(2, "0")
    if (now.getFullYear() !== dt.getFullYear()) {
        str += pad(dt.getFullYear()) + "-"
    }
    if (now.getMonth() !== dt.getMonth()) {
        str += pad(dt.getMonth()) + "-"
    }
    if (now.getDate() !== dt.getDate()) {
        str += pad(dt.getDate()) + "-"
    }
    str = str.replace(/-$/, " ")
    str += pad(dt.getHours()) + ":" + pad(dt.getMinutes()) + ":" + pad(dt.getSeconds())
    return str
}

const formatBytes = (bytes, decimals = 2) => {
    if (!+bytes) return '0 Bytes'

    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}
