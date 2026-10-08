import { addCommomParams, callAPI } from './api';
import { getUserInfo } from './login';

export type FavoriteFolder = {
  id: number;
  title: string;
  media_count: number;
};

export async function getFavorites() {
  const user = await getUserInfo();

  const api = 'https://api.bilibili.com/x/v3/fav/folder/created/list-all';
  const params = new URLSearchParams();
  params.append('up_mid', `${user.mid}`);
  addCommomParams(params);

  const url = `${api}?${params.toString()}`;

  const result = await callAPI<{
    list: FavoriteFolder[];
  }>({ url });
  return result.list;
}

export const MediaAttr = Object.freeze({
  Normal: 0, // 正常。表示该内容状态正常，可以正常访问和播放
  Invalid: 1, // 表示该内容已失效，通常是因为原视频被删除、设为私密或违反社区规则等
  Deleted: 9, // UP主自己删除。这也是一种失效状态，但特指由内容创作者自行删除
});

export type FavMedia = {
  title: string;
  cover: string;
  intro: string; // desc
  page: number;
  duration: number;
  upper: {
    mid: bigint;
    name: string;
    face: string;
  };
  attr: number;
  cnt_info: {
    collect: number; // 收藏数
    play: number; // 播放次数
    danmaku: number; // 弹幕数
  };
  pubtime: number; // 发布时间
  fav_time: number; // 收藏时间
  bvid: string;
  ugc?: {
    first_cid: bigint;
  };
};

export type FavResourceResult = {
  medias: FavMedia[];
  has_more: boolean;
};

export async function getFavResources(
  media_id: number,
  page_num: number,
  page_size: number = 40,
): Promise<FavResourceResult> {
  const api = 'https://api.bilibili.com/x/v3/fav/resource/list';
  const params = new URLSearchParams();
  params.append('media_id', `${media_id}`);
  params.append('pn', `${page_num}`);
  params.append('ps', `${page_size}`);
  params.append('order', 'mtime');
  params.append('platform', 'web');
  addCommomParams(params);

  const url = `${api}?${params.toString()}`;

  return await callAPI<FavResourceResult>({ url });
}

// 追更：我追的合集
// export async function getCollectedFolders() {}
