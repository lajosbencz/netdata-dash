const defer = () => {
    const bag = {}
    return Object.assign(
        new Promise((resolve, reject) => Object.assign(bag, { resolve, reject })),
        bag
    )
}

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

class Chart {
    static DefaultOptions = {
        broker: 'wss://localhost:16666/ws/',
        realm: 'netdata',
        theme: 'dark',
        type: 'sparkline',
        host: 'localhost',
        port: 19999,
        min: undefined,
        max: undefined,
        stack: undefined,
        smooth: true,
        symbol: 'none',
        after: -300,
        before: 0,
        group: '',
    }
    static _ResizeCallbacks = []
    static _ResizeListening = false
    static _ResizeListener = () => {
        Chart._ResizeCallbacks.forEach(cb => {
            try {
                cb()
            } catch (e) {
                console.error(e)
            }
        })
    }
    static ResizeListen(callback) {
        Chart._ResizeCallbacks.push(callback)
        if (!Chart._ResizeListening) {
            window.addEventListener('resize', Chart._ResizeListener)
            Chart._ResizeListening = true
        }
    }
    static FormatDate(dt) {
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

    static AllCharts = []
    static Create(el) {
        const metric = el.dataset.netdata
        const options = {}
        const optBool = (v) => {
            if (v === 'true' || v === '1') {
                v = true
            } else if (v === 'false' || v === '0') {
                v = false
            }
            return v
        }
        for (const k in Chart.DefaultOptions) {
            if (el.dataset.hasOwnProperty(k)) {
                let v = el.dataset[k]
                switch (k) {
                    case 'stack':
                    case 'smooth':
                        v = optBool(v)
                        break
                    case 'min':
                    case 'max':
                        v = parseFloat(v)
                        break
                    case 'port':
                    case 'after':
                    case 'before':
                        v = parseInt(v)
                        break
                }
                options[k] = v
            } else {
                options[k] = Chart.DefaultOptions[k]
            }
        }
        switch (options.type) {
            default:
            case 'sparkline':
                return new SparklineChart(el, metric, options)
            case 'sparkbar':
                return new SparkbarChart(el, metric, options)
            case 'line':
                return new LineChart(el, metric, options)
            case 'pie':
                return new PieChart(el, metric, options)
        }
    }

    static Highlight(obj, dataIndex, seriesIndex) {
        if (obj.group.length < 1) {
            return
        }
        obj.is_highlighted = true
        Chart.AllCharts.forEach(c => {
            if (obj.chart.id !== c.chart.id && obj.group === c.group && !c.is_highlighted) {
                c.is_highlighted = true
                c.chart.dispatchAction({ type: 'highlight', dataIndex, seriesIndex })
                c.chart.dispatchAction({ type: 'showTip', dataIndex, seriesIndex })
            }
        })
    }

    static Downplay(obj) {
        if (obj.group.length < 1) {
            return
        }
        obj.is_highlighted = false
        Chart.AllCharts.forEach(c => {
            if (obj.chart.id !== c.chart.id && obj.group === c.group && c.is_highlighted) {
                c.is_highlighted = false
                c.chart.dispatchAction({ type: 'downplay' })
                c.chart.dispatchAction({ type: 'hideTip' })
            }
        })
    }

    constructor(el, metric, options) {
        const nodeDataHost = el.closest('[data-host]')
        if (nodeDataHost) {
            options.host = nodeDataHost.dataset.host
            if (nodeDataHost.dataset.hasOwnProperty('port')) {
                options.port = parseInt(nodeDataHost.dataset.port)
            }
        }
        this.disposeCallbacks = []
        this.el = el
        this.metric = metric
        this.options = mergeDeep({}, Chart.DefaultOptions, options)
        this.type = this.options.type
        this.host = this.options.host + ":" + this.options.port
        this.group = this.options.group
        this.data = []
        this.chart = echarts.init(this.el, this.options.theme, { renderer: 'canvas' })
        Chart.AllCharts.push(this)
        const chartOptions = {
            backgroundColor: 'transparent',
            animation: false,
            yAxis: { show: true, axisLabel: { show: true, }, axisLine: { show: false, }, splitLine: { show: false, }, },
            xAxis: { show: false, },
            axisPointer: {
                triggerTooltip: false,
            },
        }
        for (const k of ['min', 'max']) {
            if (this.options[k] !== undefined) {
                if (!chartOptions.yAxis) chartOptions.yAxis = {}
                chartOptions.yAxis[k] = parseInt(this.options[k])
            }
        }
        this.chart.setOption(chartOptions)
        this.chart.on('highlight', e => {
            let dataIndex = e.dataIndexInside
            let seriesIndex = undefined
            if (e.batch && e.batch.length > 0) {
                dataIndex = e.batch[0].dataIndexInside
                seriesIndex = e.batch[0].seriesIndex
            }
            if (dataIndex >= 0) {
                Chart.Highlight(this, dataIndex, seriesIndex)
            }
        })
        this.chart.on('downplay', () => Chart.Downplay(this))

        Chart.ResizeListen(this.chart.resize)

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
        observeNodeRemoved(el, this.dispose)
    }
    dispose() {
        for (const cb of this.disposeCallbacks) {
            cb()
        }
        Chart._ResizeCallbacks = Chart._ResizeCallbacks.filter(e => e !== this.chart.resize)
        if (Object.values(Chart._ResizeCallbacks).length < 1) {
            window.removeEventListener('resize', Chart._ResizeListener)
        }
        Chart.AllCharts = Chart.AllCharts.filter(e => e !== this)
    }
    ondispose(callback) {
        this.disposeCallbacks.push(callback)
    }

    isInViewport() {
        const rect = this.el.getBoundingClientRect()
        const h = window.innerHeight || document.documentElement.clientHeight
        const w = window.innerWidth || document.documentElement.clientWidth
        return (
            rect.top >= -h &&
            rect.left >= -w &&
            rect.bottom >= 0 &&
            rect.right >= 0
        )
    }

    // http://localhost:19999/api/v1/charts
    initData(data) {
        this.data = []
        // console.log(this.el, this.type + ':initData', data)
    }
    // http://localhost:19999/api/v1/allmetrics?format=json&filter=system.cpu
    appendData(data) {
        // console.log(this.el, this.type + ':appendData', data)
    }
}

class ValueChart extends Chart {
    processData({ labels, data }) {
        labels = labels.slice(1)
        if (data.length < 1) {
            return
        }
        data = data.slice(-1)[0].slice(1).map((v, k) => ({
            name: labels[k],
            value: v,
        }))
        this.data = data
        this.chart.setOption({
            series: {
                type: this.type,
                data: this.data,
            },
        })
    }
    initData({ labels, data }) {
        super.initData({ labels, data })
        this.processData({ labels, data })
    }
    appendData({ labels, data }) {
        super.appendData({ labels, data })
        this.processData({ labels, data })
    }
}


class SeriesChart extends Chart {
    initData({ labels, data }) {
        super.initData({ labels, data })
        this.data = [
            labels,
            ...data.reverse(),
        ]
        this.chart.setOption({
            dataset: {
                source: this.data,
            },
            series: labels.slice(1).map(() => ({
                type: this.type,
                smooth: this.options.smooth,
                symbol: this.options.symbol,
                stack: this.options.stack,
                lineStyle: {
                    width: 1,
                },
            })),
        })
    }
    appendData({ labels, data }) {
        super.appendData({ labels, data })
        this.data.splice(1, 1)
        this.data.push(data[0])
        this.chart.setOption({
            dataset: {
                source: this.data,
            },
        })
    }
}


class LineChart extends SeriesChart {
    constructor(el, metric, options) {
        options.type = 'line'
        super(el, metric, options)
        this.chart.setOption({
            legend: {
                icon: 'rect',
            },
            tooltip: {
                trigger: 'axis',
            },
            grid: {},
            xAxis: {
                type: 'category',
                axisLabel: {
                    formatter: value => Chart.FormatDate(new Date(value * 1000)),
                },
            },
            yAxis: {},
        })
    }
}


class SparkbarChart extends ValueChart {
    constructor(el, metric, options) {
        options.type = 'bar'
        super(el, metric, options)
        this.chart.setOption({
            legend: {
                show: false,
            },
            tooltip: {
                trigger: 'axis',
            },
            grid: {
                show: false,
                top: 2,
                right: 2,
                bottom: 2,
                left: 2,
            },
            xAxis: {
                show: false,
                type: 'category',
            },
            yAxis: {
                show: false,
            },
        })
    }
}


class SparklineChart extends SeriesChart {
    constructor(el, metric, options) {
        options.type = 'line'
        super(el, metric, options)
        this.chart.setOption({
            legend: {
                show: false,
            },
            tooltip: {
                trigger: 'axis',
            },
            grid: {
                show: false,
                top: 2,
                right: 2,
                bottom: 2,
                left: 2,
            },
            xAxis: {
                show: false,
                type: 'category',
            },
            yAxis: {
                show: false,
            },
        })
    }
}


class PieChart extends ValueChart {
    constructor(el, metric, options) {
        options.type = 'pie'
        super(el, metric, options)
        this.chart.setOption({
            tooltip: {
                trigger: 'item',
            },
            legend: {
                top: '5%',
                left: 'center',
                icon: 'circle',
            },
            series: [
                {
                    name: this.metric,
                    type: this.type,
                    clockwise: false,
                    radius: ['40%', '70%'],
                    avoidLabelOverlap: false,
                    label: {
                        show: false,
                        position: 'center'
                    },
                    emphasis: {
                        label: {
                            show: true,
                            fontSize: 40,
                            fontWeight: 'bold'
                        }
                    },
                    labelLine: {
                        show: false
                    },
                },
            ],
        })
    }
}

const genTestData = (n) => {
    const t = Math.floor(+new Date() / 1000)
    const d = []
    for (let i = 0; i < n; i++) {
        const a = Math.random() * 40
        const b = Math.random() * 40
        const c = 100 - a - b
        d.push([t - i, a, b, c])
    }
    return d
}

const createChart = (ab) => async (el) => {
    const chart = Chart.Create(el)
    const params = { host: chart.host, metric: chart.metric }
    const { kwargs: { data: chartData } } = await ab.call('chart.data', [], { ...params, after: chart.options.after })
    chart.initData(chartData)
    const { kwargs: { topic } } = await ab.call('chart.topic', [], params)
    const sub = await ab.subscribe(topic, (args, { data: { last_updated, dimensions } }) => {
        const labels = []
        const data = []
        if (chart.data[0].hasOwnProperty('name')) {
            labels.push('time')
            data.push(last_updated)
            Object.values(chart.data).forEach(e => {
                labels.push(e.name)
                data.push(dimensions[e.name].value)
            })
        } else {
            chart.data[0].forEach(t => {
                labels.push(t)
                if (t === 'time') {
                    data.push(last_updated)
                } else {
                    data.push(dimensions[t].value)
                }
            })
        }
        chart.appendData({ labels, data: [data] })
    })
    chart.ondispose(() => {
        console.log('disposing', sub)
        ab.unsubscribe(sub)
    })
}

window.addEventListener('load', async () => {
    const deferSession = defer()
    const abConnection = new autobahn.Connection({ url: Chart.DefaultOptions.broker, realm: Chart.DefaultOptions.realm })
    abConnection.onopen = session => {
        deferSession.resolve(session)
    }
    abConnection.open()
    const abSession = await deferSession
    document.querySelectorAll('body [data-netdata]').forEach(createChart(abSession))

})
