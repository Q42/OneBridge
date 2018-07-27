export interface IApi {
    host: string;
    ws?: WebSocket;
    send: (str: string) => Promise<void>;
    receiveOnce: (fn: (msg: any) => boolean) => Promise<any>;
}

export function makeApi(host: string): IApi {
    let backoff = 0;
    const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${scheme}//${host}/ws`;
    let ws = new WebSocket(url);
    // tslint:disable:only-arrow-functions
    // tslint:disable:no-console
    ws.addEventListener("close, error, message, open", function() { console.log(arguments); });
    // tslint:disable:no-console
    ws.addEventListener("open", () => next());
    ws.addEventListener("close", async () => {
        await sleep(backoff);
        ws = new WebSocket(url);
        backoff += 1000;
    })

    const sendQueue: Array<{payload: string, done: IDefer<void> }> = [];
    const next = () => {
        if (!sendQueue.length) { return; }
        const { payload, done } = sendQueue.shift()!;
        try {
            ws.send(payload);
            done.resolve(undefined);
        } catch (e) {
            done.reject(e);
        }
        next();
    }

    return {
        host,
        ws,
        async send(str: string) {
            sendQueue.push({ payload: str, done: defer() });
            if (sendQueue.length === 1 && ws.readyState === WebSocket.OPEN) {
                next();
            }
        },
        async receiveOnce(matcher) {
            const d = defer()
            const handler = (msg: any) => {
                try {
                    msg = JSON.parse(msg.data);
                    if (matcher(msg)) {
                        d.resolve(msg);
                        ws.removeEventListener("message", handler);
                    }
                } catch (err) { /* ignore */ }
            }
            ws.addEventListener("message", handler);
            setTimeout(() => {
                ws.removeEventListener("message", handler);
                d.reject("timeout");
            }, 5000);
            return d.promise;
        }
    }
}

interface IDefer<T> {
    resolve: (t: T) => undefined;
    reject: (reason: any) => undefined;
    promise: Promise<T>;
}

function defer<T>() {
    const d: any = {};
    const p = new Promise<T>((r, e) => { d.resolve = r; d.reject = e; })
    return { ...d, promise: p } as IDefer<T>;
}

function sleep(ms: number) {
    return new Promise((c) => setTimeout(c, ms));
}