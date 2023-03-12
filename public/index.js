
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

const requireTag = (el, src, integrity, appendToBody) => {
    const id = src.toString().hashCode()
    if (document.getElementById(id) !== null) {
        return
    }
    el.setAttribute('id', id);
    el.setAttribute('referrerpolicy', 'no-referrer')
    el.setAttribute('crossorigin', 'anonymous');
    if (integrity) {
        el.setAttribute('integrity', integrity)
    }
    (document.getElementsByTagName(!appendToBody ? 'head' : 'body')[0]).appendChild(el)
}

const requireJs = (src, integrity, appendToBody) => {
    const d = defer()
    const el = document.createElement('script')
    el.setAttribute('type', 'text/javascript')
    el.setAttribute('src', src)
    el.addEventListener('load', d.resolve)
    requireTag(el, src, integrity, appendToBody)
    return d
}

const requireStyle = (src, integrity) => {
    const el = document.createElement('link')
    el.setAttribute('href', src)
    el.setAttribute('rel', 'stylesheet')
    requireTag(el, src, integrity, false)
}

const NETDATA_DEFAULTS = {
    autoStart: true,
    WAMP_ADDRESS: "ws://localhost:9301",
    WAMP_REALM: "netdata",
    FORMAT_DATETIME: (dt) => {
        return (new Date(dt)).toISOString().substring(0, 19)
    },
};

NETDATA = Object.assign({}, NETDATA_DEFAULTS, window.NETDATA || {});

NETDATA.start = async function () {

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

        const createChart = async el => {
        }

        // const createChartDygraph = async el => {
        //     const wamp = await deferWampSession
        //     const host = el.dataset.host || 'localhost:19999'
        //     const chart = el.dataset.chart || 'system.cpu'
        //     const chartType = el.dataset.type || 'line'
        //     const height = parseInt(el.dataset.height || 350)
        //     const { kwargs: { data: { labels, data } } } = await wamp.call('data', [], {
        //         host,
        //         chart,
        //         before: 0,
        //         after: -60,
        //     })
        //     console.log({ labels, data })

        //     const graph = new Dygraph(el, data.reverse(), {
        //         showRoller: true,
        //         gridLineColor: 'rgba(255,255,255,.05)',
        //         labels,
        //     })

        //     return {
        //         graph,
        //         host,
        //         chart,
        //         labels,
        //         data,
        //     }
        // }

        // x-netdata
        //Alpine.directive('netdata', (el, { value, modifiers, expression }, { Alpine, evaluate, effect, cleanup }) => {

        Alpine.directive('netdata', async (el, a, b) => {
            // console.log('x-netdata', a, b)
            const wamp = await deferWampSession
            let { graph, labels, data: oldData, host, chart } = await createChart(el)
            const sub = await wamp.subscribe('data.' + chart, (args, { data: newData }) => {
                oldData = [...(oldData.slice(1)), newData]
                console.log(oldData)
                graph.updateOptions({ 'file': oldData })
            })
            console.log({ sub, graph, oldData, host, chart })
        })

    })

    document.addEventListener('alpine:initialized', async () => {
        Alpine.effect(wampConnect)
        wampConnect()
    })
}

    ; (async () => {
        requireStyle('//cdnjs.cloudflare.com/ajax/libs/picocss/1.5.7/pico.min.css',
            'sha512-1VnpjjanhjGWRcbZCUKqh1KbNIGAd8aqsokcHUNlBFM3CfAUasd7D0h1luMzyS01W74K4zUZq7GZnj3yoGYEFQ==')
        requireStyle('//cdnjs.cloudflare.com/ajax/libs/font-awesome/6.3.0/css/all.min.css',
            'sha512-SzlrxWUlpfuzQ+pcUCosxcglQRNAq/DZjVsC0lE40xsADsfeQoEypE+enwcOiGjk/bSuGGKHEyjSoQ1zVisanQ==')
        await Promise.all([
            requireJs('//cdnjs.cloudflare.com/ajax/libs/autobahn/22.10.1/autobahn.min.js',
                'sha512-NV3SvHAZNmkfgYNYbooVfXPHOXSxozk0TJALPt9J2xk1cVwp0YnTw5k3W6IClirda/A9DspvjeBqxmgPvdus+w=='),
            // requireJs('//cdnjs.cloudflare.com/ajax/libs/apexcharts/3.37.1/apexcharts.min.js'),
            // requireJs('//cdnjs.cloudflare.com/ajax/libs/dygraph/2.2.1/dygraph.min.js'),
            requireJs('//cdnjs.cloudflare.com/ajax/libs/echarts/5.4.1/echarts.min.js',
                'sha512-OTbGFYPLe3jhy4bUwbB8nls0TFgz10kn0TLkmyA+l3FyivDs31zsXCjOis7YGDtE2Jsy0+fzW+3/OVoPVujPmQ=='),
        ])
        NETDATA.start();
        requireJs('//cdnjs.cloudflare.com/ajax/libs/alpinejs/3.12.0/cdn.js',
            'sha512-6pVa1JFPLsAVloI/eZXmbDkCWYVB3Y8ODVA2gVUIowY2laRHAYaYPE1f4KjSvPwNimMmGo4MvteDc3JZEjEikA==')
    })()
