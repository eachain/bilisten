import * as z from 'zod';
import { proxyAPI } from './backend';
import type { Request } from './backend';

export const WebIfaceEndpoint = 'https://api.bilibili.com/x/web-interface';

export class BiliError extends Error {
  code: number;
  message: string;

  constructor(code: number, message: string) {
    super(`bilibili error ${code}: ${message}`);
    this.code = code;
    this.message = message;
  }
}

type Response = {
  code: number;
  message: string;
  data: any;
};

export async function callAPI<T>(req: Request): Promise<T> {
  const res = await proxyAPI(req);
  if (res.status_code !== 200) {
    throw Error(`bilibili server http ${res.status}`);
  }

  const response = z.custom<Response>().parse(res.body);
  if (response.code !== 0) {
    throw new BiliError(response.code, response.message);
  }

  return z.custom<T>().parse(response.data);
}

export function addCommomParams(params: URLSearchParams) {
  params.append('web_location', '333.1007');
  params.append(
    'x-bili-locale-json',
    '{"c_locale":{"language":"zh","script":"Hans"},"always_translate":true}',
  );
  params.append('x-bili-redirect', '1');
  params.append(
    'x-bili-device-req-json',
    '{"platform":"web","device":"pc","spmid":"333.1007","mobi_app":"web_cn"}',
  );
}
