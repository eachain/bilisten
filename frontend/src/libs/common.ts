import type { Music } from './player';
import { getSongView } from './search';

// MARK: play music by bvid

export async function playMusicByBvid(
  bvid: string,
  onlyOne: (m: Music) => void,
  multi: (ms: Music[]) => void,
) {
  const dash = await getSongView(bvid);
  if (dash.pages.length === 0) {
    throw Error('no pages found');
  }

  if (dash.pages.length === 1) {
    onlyOne({
      bvid: dash.bvid,
      cid: dash.cid,
      title: dash.title,
      desc: dash.desc,
      duration: dash.pages[0].duration,
      author: dash.owner.name,
      mid: `${dash.owner.mid}`,
      pic: dash.pic,
      page: dash.pages[0].page,
      play: dash.stat.view,
      favorites: dash.stat.favorite,
      danmaku: dash.stat.danmaku,
      like: dash.stat.like,
      pubdate: dash.pubdate,
    });
    return;
  }

  const musics: Music[] = [];
  for (const page of dash.pages) {
    musics.push({
      bvid: dash.bvid,
      cid: page.cid,
      title: page.part,
      desc: dash.desc,
      duration: page.duration,
      author: dash.owner.name,
      mid: `${dash.owner.mid}`,
      page: page.page,
      pic: page.first_frame ?? dash.pic,
      pubdate: dash.pubdate,
    });
  }
  multi(musics);
}
