

const NETDATA_DEFAULTS = {
    autoStart: true,
    WAMP_ADDRESS: "wss://" + location.host + "/ws/",
    WAMP_REALM: "netdata",
    // FORMAT_DATETIME: (dt) => {
    //     return (new Date(dt)).toISOString().substring(0, 19)
    // },
};

window.NETDATA = Object.assign({}, NETDATA_DEFAULTS, window.NETDATA || {});

; (async () => {

    window.NETDATA.start = async function () { }
    window.NETDATA.stop = async function () { }
    window.NETDATA.restart = async function () {
        try { await window.NETDATA.stop() }
        catch (e) { }
        await window.NETDATA.start()
    }

    let deferWampSession = defer()

    let wampConn = undefined
    const wampOpen = (onopen) => {
        wampConn = new autobahn.Connection({ url: NETDATA.WAMP_ADDRESS, realm: NETDATA.WAMP_REALM });
        wampConn.onopen = onopen;
        wampConn.open();
    }
    const wampDisconnect = () => {
        if (wampConn) {
            wampConn.close();
            try {
                deferWampSession.reject()
            } catch (e) {
                console.error(e)
            }
            deferWampSession = defer()
        }
        wampConn = undefined;
    }
    const wampConnect = () => {
        const running = Alpine.store('app').running
        if (running && wampConn === undefined) {
            wampOpen(session => {
                deferWampSession.resolve(session)
            })
        }
        if (!running && wampConn !== undefined) {
            wampDisconnect();
        }
    }

    document.addEventListener('alpine:init', () => {

        // x-log
        Alpine.directive('log', (el, { expression }, { evaluate }) => {
            console.log(evaluate(expression))
        })

        // $dateTimeFormat(Date)
        Alpine.magic('dateTimeFormat', () => dt => {
            return (new Date(dt)).toISOString().substring(0, 19)
        })

        Alpine.store('app', {
            running: NETDATA.autoStart,
            duration: 600,
            dateFrom: undefined,
            dateUntil: undefined,
            hosts: [],
        })
        const appStore = Alpine.store('app')
        appStore.dateFrom = formatDate(new Date((Date.now() - appStore.duration)))
        appStore.dateUntil = formatDate(new Date(Date.now()))

        Alpine.data('app', () => ({
            init() {
            },

            runStart: {
                ['x-if']() { return !appStore.running },
            },

            runStartClick: {
                ['@click']() { appStore.running = true },
            },

            runStop: {
                ['x-if']() { return appStore.running },
            },

            runStopClick: {
                ['@click']() { appStore.running = false },
            },
        }))

        const chartOptionsFromEl = (el) => {
            const options = {}
            const defaults = {
                metric: undefined,
                host: 'localhost',
                type: 'line',
                after: -600,
                before: 0,
                stack: undefined,
                smooth: undefined,
                min: undefined,
                max: undefined,
            }
            for (const k in defaults) {
                if (el.dataset.hasOwnProperty(k)) {
                    let v = el.dataset[k]
                    switch (k) {
                        case 'stack':
                        case 'smooth':
                            v = !(['0', 'false'].includes(v))
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
                    options[k] = defaults[k]
                }
            }
            const elp = el.closest("[data-host]")
            if (elp) {
                options.host = elp.dataset.host
            }
            return options
        }

        const createChart = async (wamp, el) => {
            const options = chartOptionsFromEl(el)
            const topic = 'chart.data._.' + options.host + '._.' + options.metric
            const params = { metric: options.metric, after: options.after, before: options.before }
            const { kwargs: { labels, data } } = await wamp.call('chart.data._.' + options.host, [], params)
            const rows = [labels, ...data]
            const graph = bb.generate({
                bindto: el,
                transition: {
                    duration: 0,
                },
                tooltip: {
                    linked: true,
                },
                line: {
                    point: false,
                },
                data: {
                    x: "time",
                    type: "line",
                    rows,
                },
                axis: {
                    x: {
                        type: "timeseries",
                        tick: {
                            culling: {
                                max: 5
                            },
                            format(x) {
                                return smartFormatDate(new Date(x * 1000))
                            },
                        },
                    },
                },
            })
            const sub = await wamp.subscribe(topic, (args, { metricName, metricData: { last_updated, dimensions } }) => {
                dimensions.time = { name: "time", value: last_updated }
                const row = []
                for (const l of labels) {
                    row.push(dimensions[l].value)
                }
                graph.flow({
                    rows: [
                        labels,
                        row,
                    ],
                })

            })
            observeNodeRemoved(el, () => {
                console.log('disposing', sub)
                wamp.unsubscribe(sub)
            })
            return { graph, labels, options }
        }

        const createEChart = async (wamp, el) => {
            const chart = Chart.Create(el)
            const topic = 'chart.data._.' + chart.host + '._.' + chart.metric
            const params = { metric: chart.metric, after: chart.options.after }
            const { kwargs: chartData } = await wamp.call('chart.data._.' + chart.host, [], params)
            chart.initData(chartData)
            const sub = await wamp.subscribe(topic, (args, { metricName, metricData: { last_updated, dimensions } }) => {
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
                wamp.unsubscribe(sub)
            })
        }

        // x-netdata
        //Alpine.directive('netdata', (el, { value, modifiers, expression }, { Alpine, evaluate, effect, cleanup }) => {

        Alpine.directive('netdata', async (el, a, b) => {
            // console.log('x-netdata', a, b)
            const wamp = await deferWampSession
            try {
                const { graph, labels, options: { metric, host } } = await createChart(wamp, el)
                console.log({ graph, labels, host, metric })
            } catch (e) {
                console.error(e)
            }
        })

    })

    document.addEventListener('alpine:initialized', async () => {
        Alpine.effect(wampConnect)
        wampConnect()

        const wamp = await deferWampSession
        const appStore = Alpine.store('app')
        wamp.subscribe('host.list', (args) => {
            appStore.hosts = args
        })
        const { kwargs: { list } } = await wamp.call('host.list')
        appStore.hosts = list
    })

})()