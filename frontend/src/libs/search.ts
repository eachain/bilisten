import { ref } from 'vue';
import { callAPI, WebIfaceEndpoint } from './api';
import { logger } from './backend';

export const currentKeyword = ref('');
export const totalResultCount = ref(0);

export type HitItem = {
  author: string;
  mid: bigint;
  bvid: string;
  title: string;
  duration: string; // 时长 M:S 格式
  description: string;
  pic: string;
  play: number; // 播放次数
  favorites: number; // 收藏数
  danmaku: number; // 弹幕数
  like: number; // 点赞
  senddate: number; // unix timestamp 发布时间
};

export type SearchResult = {
  numResults: number;
  numPages: number;
  result: HitItem[];
};

export async function search(
  keyword: string,
  page: number,
): Promise<SearchResult> {
  logger.info(`search '${keyword}' page ${page}`);
  const params = new URLSearchParams();
  params.append('search_type', 'video');
  params.append('keyword', keyword);
  params.append('page', `${page}`);
  params.append('page_size', `20`);

  const api = `${WebIfaceEndpoint}/search/type`;
  const url = `${api}?${params.toString()}`;
  const result = await callAPI<SearchResult>({ url });
  return result;
}

export const searchResults = ref<HitItem[]>([]);
export const hasMore = ref(false);

export let nextPage: () => void = () => {};

export function setNextPage(fn: () => void) {
  nextPage = fn;
}

export function formatCount(n: number): string {
  if (n < 10000) {
    return `${n}`;
  }
  if (n < 10000_0000) {
    let w = (n / 10000).toFixed(1);
    if (w.endsWith('.0')) {
      w = w.slice(0, w.length - 2);
    }
    return w + '万';
  }
  let w = (n / 10000_0000).toFixed(1);
  if (w.endsWith('.0')) {
    w = w.slice(0, w.length - 2);
  }
  return w + '亿';
}

export function formatSendDate(sd: number): string {
  const date = new Date(sd * 1000);
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${year}年${month}月${day}日`;
}

export function parseDuration(dur: string): number {
  if (dur === '') {
    return 0;
  }
  return dur
    .split(':')
    .map((p) => Number(p))
    .reduce((prev, curr) => prev * 60 + curr, 0);
}

// 搜索的结果是'M:S'格式，转为'HH:MM:SS'格式
export function reFormatDuration(dur: string): string {
  if (dur === '') {
    return '--:--';
  }
  const parts = dur.split(':');
  if (parts.length === 1) {
    return '00:' + dur;
  }
  if (parts.length === 2) {
    parts[1] = parts[1].padStart(2, '0');
    let minutes = Number(parts[0]);
    if (minutes < 60) {
      parts[0] = parts[0].padStart(2, '0');
      return parts.join(':');
    }
    const hours = Math.floor(minutes / 60);
    minutes = minutes % 60;
    return [
      `${hours}`.padStart(2, '0'),
      `${minutes}`.padStart(2, '0'),
      parts[1],
    ].join(':');
  }
  return parts.map((p) => p.padStart(2, '0')).join(':');
}

export type SuggestResults = {
  term: string;
  value: string;
  name: string; // html
}[];

export async function suggest(keyword: string): Promise<SuggestResults> {
  logger.info(`suggest '${keyword}'`);
  const params = new URLSearchParams();
  params.append('term', keyword);

  const api = `${WebIfaceEndpoint}/suggest`;
  const url = `${api}?${params.toString()}`;
  const result = await callAPI<{
    result: { tag: SuggestResults };
  }>({ url });
  return result.result.tag;
}

export type SongView = {
  bvid: string;
  pic: string;
  title: string;
  desc: string;
  pubdate: number;
  owner: {
    mid: bigint;
    name: string;
    face: string; // http://xxx.jpg
  };
  stat: {
    view: number; // 播放次数
    danmaku: number; // 弹幕数
    favorite: number; // 收藏数
    like: number; // 点赞
    reply: number; // 评论数
    coin: number; // 投币数
    share: number; // 分享数
  };
  cid: bigint;
  pages: {
    cid: bigint;
    page: number;
    part: string;
    duration: number;
    first_frame: string; // http://xxx.jpg
  }[];
};

export async function getSongView(bvid: string): Promise<SongView> {
  logger.info(`view '${bvid}'`);
  const params = new URLSearchParams();
  params.append('bvid', bvid);

  const api = `${WebIfaceEndpoint}/view`;
  const url = `${api}?${params.toString()}`;
  return await callAPI<SongView>({ url });
}

export type SongPlayURLDash = {
  duration: number;
  audio: {
    id: number; // SoundQuality
    base_url: string;
    backup_url: string[];
    bandwidth: number;
    mime_type: string; // "audio/mp4"
    codecs: string; // "mp4a.40.2"
  }[];
};

export async function getPlayURL(
  bvid: string,
  cid: bigint,
): Promise<SongPlayURLDash> {
  logger.info(`get playurl '${bvid} ${cid}'`);
  const params = new URLSearchParams();
  params.append('bvid', bvid);
  params.append('cid', `${cid}`);
  params.append('fnval', `${16 | 256}`); // 16: dash, 256: 杜比音频

  const url = `https://api.bilibili.com/x/player/playurl?${params.toString()}`;
  const result = await callAPI<{
    dash: SongPlayURLDash;
  }>({ url });
  return result.dash;
}
