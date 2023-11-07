import {rest} from "msw";
import {setupServer} from "msw/node";
import {afterAll, afterEach, beforeAll} from "vitest";

type restMethod =
    'all' |
    'head' |
    'get' |
    'post' |
    'put' |
    'delete' |
    'patch' |
    'options'

export class CreateServerConfig {
    method?: restMethod
    path: string
    res: ({req, res, context}:{req: any, res: any, context: any}) => any;


    constructor({method, path, res}:
                    {
                        method?: restMethod,
                        path: string,
                        res: ({req, res, context}:{req: any, res: any, context: any}) => any;
                    }) {
        this.method = method;
        this.path = path;
        this.res = res;
    }
}

export const createServer = (handlerConfig: CreateServerConfig[]) => {
    const handlers = handlerConfig.map(config => {
        return rest[config.method || 'get'](config.path, (req, res, context) => {
            return res(
                context.json(
                    config.res({req, res, context})
                )
            )
        })
    })
    const server = setupServer(...handlers)
    beforeAll(() => server.listen())
    afterEach(() => server.resetHandlers())
    afterAll(() => server.close())
}
