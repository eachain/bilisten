import * as z from 'zod';

// const BackendURL = 'http://127.0.0.1:65511';
const BackendURL = '';

export type Request = {
  method?: string;
  url: string;
  csrf?: string;
  headers?: [string, string][];
  body?: any;
  timeout?: number;
};

export type Response = {
  status_code: number;
  status: string;
  headers: [string, string][];
  body: any;
};

const ResponseZ = z.custom<Response>();

export async function proxyAPI(req: Request): Promise<Response> {
  req.timeout = req.timeout ?? 10000;
  const resp = await fetch(BackendURL + '/api', {
    method: 'POST',
    body: JSON.stringify(req),
  });
  return ResponseZ.parse(JSON.parse(await resp.text()));
}

export function thumbnail(url: string, width: number = 100): string {
  const ext = '.webp';

  if (!url.startsWith('http')) {
    url = 'http:' + url;
  }
  if (width > 0) {
    url += `@${width}w${ext}`;
  } else {
    url += `@${ext}`;
  }
  return url;
}

export function mustHTTPS(url: string): string {
  if (url.startsWith('http://')) {
    url = 'https' + url.slice(4);
  }
  return url;
}

export function proxyImage(url: string, width: number = 100): string {
  url = thumbnail(url, width);
  const params = new URLSearchParams();
  params.append('url', url);
  return `${BackendURL}/image.webp?${params.toString()}`;
}

export type Quality = 'flac' | 'dolby' | 'high' | 'mid' | 'low';

export const SoundQualities: {
  id: 30216 | 30232 | 30280 | 30250 | 30251;
  q: Quality;
  name: '流畅' | '标准' | '高清' | '杜比' | '无损';
  desc: string;
}[] = [
  {
    id: 30216,
    q: 'low',
    name: '流畅',
    desc: '高频细节损失明显，但节省流量',
  },
  {
    id: 30232,
    q: 'mid',
    name: '标准',
    desc: '人声和乐器基本轮廓得以保留',
  },
  {
    id: 30280,
    q: 'high',
    name: '高清',
    desc: '多数设备下与无损的差异不易察觉',
  },
  {
    id: 30250,
    q: 'dolby',
    name: '杜比',
    desc: '必须登录，需设备支持，三维空间环绕的沉浸式听感',
  },
  {
    id: 30251,
    q: 'flac',
    name: '无损',
    desc: '必须登录，Hi-Res还原录音室级别的听感',
  },
];

export function getQualityName(idOrQ: string): string {
  return (
    SoundQualities.filter(
      (sq) => String(sq.id) === idOrQ || sq.q === idOrQ,
    )?.[0]?.name ?? idOrQ
  );
}

export function proxyMPD(
  bvid: string,
  cid: bigint,
  title: string,
  pic: string,
  author: string,
  mid: string,
  page?: number,
  quality: string = 'high,flac,dolby,mid,low',
): string {
  const params = new URLSearchParams();
  params.append('title', title);
  params.append('pic', pic);
  params.append('author', author);
  params.append('mid', mid);
  if (quality === '') {
    quality = 'high,flac,dolby,mid,low';
  }
  params.append('q', quality);
  if (page) {
    params.append('page', `${page}`);
  }
  return `${BackendURL}/mpd/${bvid}/${cid}.mpd?${params.toString()}`;
}

export async function getAudioQualities(
  bvid: string,
  cid: bigint,
): Promise<string[]> {
  const params = new URLSearchParams();
  params.append('bvid', bvid);
  params.append('cid', `${cid}`);
  const url = `${BackendURL}/cache/audio/q?${params.toString()}`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function getCacheSize(
  type: 'audio' | 'image',
): Promise<`${number}`> {
  const url = `${BackendURL}/cache/${type}/size`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function clearCache(type: 'audio' | 'image') {
  const url = `${BackendURL}/cache/${type}/clear`;
  const resp = await fetch(url);
  await resp.text();
}

export async function getSearchHistory(): Promise<string[]> {
  const url = `${BackendURL}/store/search`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function setSearchKeyword(keyword: string): Promise<string[]> {
  const params = new URLSearchParams();
  params.append('keyword', keyword);
  const url = `${BackendURL}/store/search?${params.toString()}`;
  const resp = await fetch(url, { method: 'POST' });
  return await resp.json();
}

export async function delSearchKeyword(keyword: string): Promise<string[]> {
  const params = new URLSearchParams();
  params.append('keyword', keyword);
  const url = `${BackendURL}/store/search?${params.toString()}`;
  const resp = await fetch(url, { method: 'DELETE' });
  return await resp.json();
}

export async function clearAllSearchKeywords(): Promise<string[]> {
  const params = new URLSearchParams();
  params.append('clear', 'true');
  const url = `${BackendURL}/store/search?${params.toString()}`;
  const resp = await fetch(url, { method: 'DELETE' });
  return await resp.json();
}

export async function getPlaylist(): Promise<any[]> {
  const url = `${BackendURL}/store/play/list`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function updatePlaylist(list: any[]) {
  const url = `${BackendURL}/store/play/list`;
  const resp = await fetch(url, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(list),
  });
  await resp.text();
}

export async function getPlayState(): Promise<{ mode: string; index: string }> {
  const url = `${BackendURL}/store/play/state`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function setPlayState({
  mode = '',
  index = '',
}: {
  mode?: string;
  index?: string;
}) {
  const params = new URLSearchParams();
  params.append('mode', mode);
  params.append('index', index);
  const url = `${BackendURL}/store/play/state?${params.toString()}`;
  const resp = await fetch(url, { method: 'POST' });
  await resp.text();
}

export async function getSettings(): Promise<{
  buffer: number;
  audio: number;
  image: number;
  quality: string;
}> {
  const url = `${BackendURL}/store/settings`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function setSettings(
  key: 'buffer' | 'audio' | 'image' | 'quality',
  value: string,
): Promise<number | string> {
  const params = new URLSearchParams();
  params.append('key', key);
  params.append('value', value);
  const url = `${BackendURL}/store/settings?${params.toString()}`;
  const resp = await fetch(url, { method: 'POST' });
  return await resp.json();
}

export async function isLogined(): Promise<boolean> {
  const url = `${BackendURL}/store/cookie/logined`;
  const resp = await fetch(url);
  return await resp.json();
}

export async function longPollLoginEvent(): Promise<boolean> {
  const url = `${BackendURL}/store/cookie/poll`;
  const resp = await fetch(url);
  return await resp.json();
}

export function videoURL(bvid: string, page?: number): string {
  const url = `https://www.bilibili.com/video/${bvid}/`;
  if (page) {
    return url + `?p=${page}`;
  }
  return url;
}

export function userSpaceURL(mid: string): string {
  return `https://space.bilibili.com/${mid}`;
}
