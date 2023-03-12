

const NETDATA_DEFAULTS = {
    autoStart: true,
    WAMP_ADDRESS: "wss://localhost:16666/ws/",
    WAMP_REALM: "netdata",
    FORMAT_DATETIME: (dt) => {
        return (new Date(dt)).toISOString().substring(0, 19)
    },
};

NETDATA = Object.assign({}, NETDATA_DEFAULTS, window.NETDATA || {});

NETDATA.start = async function () {
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
    })
    const appStore = Alpine.store('app')
    appStore.dateFrom = NETDATA.FORMAT_DATETIME(new Date((Date.now() - appStore.duration)))
    appStore.dateUntil = NETDATA.FORMAT_DATETIME(new Date(Date.now()))

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

    const createChart = async (wamp, el) => {
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
            let { graph, labels, data: oldData, host, chart } = await createChart(wamp, el)
            const sub = await wamp.subscribe('data.' + chart, (args, { data: newData }) => {
                oldData = [...(oldData.slice(1)), newData]
                console.log(oldData)
                graph.updateOptions({ 'file': oldData })
            })
            console.log({ sub, graph, oldData, host, chart })
        } catch (e) {
            console.error(e)
        }
    })

})

document.addEventListener('alpine:initialized', async () => {
    Alpine.effect(wampConnect)
    wampConnect()
})
