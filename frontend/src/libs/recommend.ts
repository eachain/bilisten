import { callAPI } from './api';

export type NewMusicInfo = {
  music_title: string;
  publish_time: string; // format: 'm-d'
  wish_count: number; // 想听
  total_vv: number; // 播放量
  cover: string;
  author: string;
  album: string;
  bvid: string;
};

export async function loadNewMusics(): Promise<NewMusicInfo[]> {
  const params = new URLSearchParams();
  params.append('plat', '2');
  params.append('web_location', '333.1007');
  const api = 'https://api.bilibili.com/x/centralization/interface/new/music';
  const url = `${api}?${params.toString()}`;
  const result = await callAPI<{
    list: NewMusicInfo[];
  }>({ url, csrf: 'csrf' });

  return result.list;
}

export type HotMusic = {
  music_title: string;
  bvid: string;
  cid: string;
  author: string;
  album: string;
  cover: string;
  total_vv: number; // 播放量
  wish_count: number; // 想听
};

export async function loadHotMusics(): Promise<HotMusic[]> {
  const params = new URLSearchParams();
  params.append('plat', '2');
  params.append('web_location', '333.1007');
  const api =
    'https://api.bilibili.com/x/centralization/interface/music/hot/rank';
  const url = `${api}?${params.toString()}`;
  const result = await callAPI<{
    list: HotMusic[];
  }>({ url, csrf: 'csrf' });

  return result.list;
}
