import * as z from 'zod';
import { ref } from 'vue';
import { MediaPlayer, Debug } from 'dashjs';
import type { MediaPlayerClass } from 'dashjs';
import {
  getPlaylist,
  getPlayState,
  getQualityName,
  getSettings,
  logger,
  mustHTTPS,
  proxyMPD,
  setPlayState,
  thumbnail,
  updatePlaylist,
} from './backend.ts';
import { noticeAlert } from './alerts.ts';
import { type AnimateStatus, setAnimateStatus } from './animate.ts';

export type Music = {
  bvid: string;
  cid: bigint;
  title: string;
  desc: string; // useless
  pic: string;
  duration: number;
  author: string;
  mid: string;
  page?: number;
  play?: number; // 播放次数
  favorites?: number; // 收藏数
  danmaku?: number; // 弹幕数
  like?: number; // 点赞
  pubdate?: number; // unix timestamp 发布时间
};

export const playmode = ref<'list' | 'repeat' | 'shuffle'>('list');
let currentPlayer: MediaPlayerClass | undefined = undefined;

// play info
export const title = ref('Bilisten');
export const cover = ref('');
export const progress = ref(0); // 0~duration
export const buffered = ref(0); // 0~duration
export const duration = ref(0); // int
export const quality = ref('音质'); // low, high, ...

export const playing = ref(false);
export const playlist = ref<Music[]>([]);
export const playindex = ref(0); // always point to the playlist[playindex];
let randlist: Music[] = [];
let randindex = 0;
let bufferTime = 30; // 预加载秒数
let qualityPrefer = ''; // 音质偏好

export async function loadPlayerState() {
  const { buffer, quality } = await getSettings();
  bufferTime = buffer;
  qualityPrefer = quality;

  playlist.value = await getPlaylist();
  const { mode, index } = await getPlayState();
  if (mode === 'list' || mode === 'repeat' || mode == 'shuffle') {
    playmode.value = mode;
  }
  playindex.value = Number(index ?? '0');
  if (playindex.value < 0 || playindex.value >= playlist.value.length) {
    playindex.value = 0;
  }

  if (playmode.value === 'shuffle' && playlist.value.length > 0) {
    randlist = [...playlist.value];
    shuffle(randlist);
    const curr = playlist.value[playindex.value];
    randindex = randlist.findIndex(
      (m) => m.bvid == curr.bvid && m.cid == curr.cid,
    );
  }

  if (playlist.value.length > 0) {
    playMusic(playlist.value[playindex.value], false);
  }
}

function resetCurrent() {
  title.value = 'Bilisten';
  cover.value = '';
  progress.value = 0; // 0~duration
  buffered.value = 0; // 0~duration
  duration.value = 0; // int
  quality.value = '音质'; // low, high, ...
  currentPlayer?.destroy();
  currentPlayer = undefined;
}

function resetList() {
  playing.value = false;
  playlist.value = [];
  playindex.value = 0;
  randlist = [];
  randindex = 0;
  updatePlaylist([]);
  setPlayState({ index: '0' });
}

export function clearCurrentPlaylist() {
  resetCurrent();
  resetList();
}

export function shuffle<T>(list: T[]) {
  for (let i = list.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [list[i], list[j]] = [list[j], list[i]];
  }
}

export function chmod() {
  if (playmode.value === 'list') {
    playmode.value = 'repeat';
    setPlayState({ mode: playmode.value });
    return;
  }

  if (playmode.value === 'repeat') {
    playmode.value = 'shuffle';
    if (playlist.value.length > 0) {
      randlist = [...playlist.value];
      shuffle(randlist);
      const curr = playlist.value[playindex.value];
      randindex = randlist.findIndex(
        (m) => m.bvid == curr.bvid && m.cid == curr.cid,
      );
    }
    setPlayState({ mode: playmode.value });
    return;
  }

  if (playmode.value === 'shuffle') {
    playmode.value = 'list';
    randlist = [];
    randindex = 0;
    setPlayState({ mode: playmode.value });
    return;
  }
}

export function playNext(autoPlay: boolean = true) {
  if (currentPlayer && playlist.value.length < 2) {
    return;
  }

  if (playmode.value === 'list' || playmode.value === 'repeat') {
    if (currentPlayer) {
      playindex.value = (playindex.value + 1) % playlist.value.length;
      playMusic(playlist.value[playindex.value], autoPlay);
    } else {
      playMusic(playlist.value[0], autoPlay);
    }
    return;
  }

  if (playmode.value === 'shuffle') {
    if (currentPlayer) {
      randindex = (randindex + 1) % randlist.length;
      playMusic(randlist[randindex], autoPlay);
    } else {
      playMusic(randlist[0], autoPlay);
    }
  }
}

export function playPrev(autoPlay: boolean = true) {
  if (currentPlayer && playlist.value.length < 2) {
    return;
  }
  if (playmode.value === 'list' || playmode.value === 'repeat') {
    if (currentPlayer) {
      playindex.value =
        (playindex.value + playlist.value.length - 1) % playlist.value.length;
      playMusic(playlist.value[playindex.value], autoPlay);
    } else {
      playMusic(playlist.value[0], autoPlay);
    }
    return;
  }

  if (playmode.value === 'shuffle') {
    if (currentPlayer) {
      randindex = (randindex + randlist.length - 1) % randlist.length;
      playMusic(randlist[randindex], autoPlay);
    } else {
      playMusic(randlist[0], autoPlay);
    }
  }
}

export function pushToPlaylist(music: Music, play: boolean = false) {
  const index = playlist.value.findIndex(
    (m) => m.bvid === music.bvid && m.cid === music.cid,
  );
  if (index >= 0) {
    return;
  }

  playlist.value.push(music);
  updatePlaylist(playlist.value);
  if (play) {
    playindex.value = playlist.value.length - 1;
  }

  if (playmode.value === 'shuffle') {
    randlist = [
      ...randlist.slice(0, randindex + 1),
      music,
      ...randlist.slice(randindex + 1),
    ];
    if (play) {
      randindex++;
    }
  }
}

export function pushMultiToPlaylist(musics: Music[]): Music[] {
  if (musics.length === 0) {
    return [];
  }
  const exists = new Map<string, boolean>();
  for (const m of playlist.value) {
    exists.set(`${m.bvid}-${m.cid}`, true);
  }

  const pushed: Music[] = [];
  for (const m of musics) {
    const key = `${m.bvid}-${m.cid}`;
    if (!exists.get(key)) {
      exists.set(key, true);
      pushed.push(m);
    }
  }
  if (pushed.length === 0) {
    return [];
  }

  playlist.value = [...playlist.value, ...pushed];
  updatePlaylist(playlist.value);

  if (playmode.value === 'shuffle') {
    const tmp = [...pushed];
    shuffle(tmp);
    randlist = [
      ...randlist.slice(0, randindex + 1),
      ...tmp,
      ...randlist.slice(randindex + 1),
    ];
  }

  return pushed;
}

export function popFromPlaylist(music: Music) {
  const index1 = playlist.value.findIndex(
    (m) => m.bvid === music.bvid && m.cid === music.cid,
  );
  if (index1 === -1) {
    return;
  }

  playlist.value = [
    ...playlist.value.slice(0, index1),
    ...playlist.value.slice(index1 + 1),
  ];
  updatePlaylist(playlist.value);

  if (playmode.value === 'shuffle') {
    const index2 = randlist.findIndex(
      (m) => m.bvid === music.bvid && m.cid === music.cid,
    );
    if (index2 !== -1) {
      randlist = [...randlist.slice(0, index2), ...randlist.slice(index2 + 1)];
      if (index2 < randindex) {
        randindex--;
      }
    }
  }

  if (index1 === playindex.value) {
    resetCurrent();
    if (playlist.value.length === 0) {
      resetList();
      return;
    }
    if (playlist.value.length === 1) {
      playindex.value = 0;
      randindex = 0;
      playMusic(playlist.value[0], playing.value);
      return;
    }
    if (playmode.value === 'shuffle') {
      randindex = (randindex + randlist.length - 1) % randlist.length;
    } else {
      playindex.value =
        (playindex.value + playlist.value.length - 1) % playlist.value.length;
    }
    playNext(playing.value);
  } else if (index1 < playindex.value) {
    playindex.value--;
    setPlayState({ index: `${playindex.value}` });
  }
}

let getAudioPlayer: () => HTMLMediaElement;

export function setAudioPlayer(get: () => HTMLMediaElement) {
  getAudioPlayer = get;
}

export function playNow(
  music: Music,
  set: (status: AnimateStatus) => void = () => {},
) {
  if (isCurrentPlaying(music.bvid, music.cid)) {
    currentPlayer?.play();
    return;
  }
  playMusic(music, true, set);
}

const DashManifestInfoZ = z.object({
  mediaPresentationDuration: z.number(),
  Period: z
    .object({
      AdaptationSet: z
        .object({
          Representation: z
            .object({
              id: z.templateLiteral([z.number()]),
            })
            .array(),
        })
        .array(),
    })
    .array(),
});

let onPlayCallback = (_: Music) => {};

export function setOnPlay(cb: (music: Music) => void) {
  onPlayCallback = cb;
}

export function updateBufferTime(buffer: number) {
  bufferTime = buffer;
}

export function updateQualityPrefer(quality: string) {
  qualityPrefer = quality;
}

let failedCount = 0;

function playMusic(
  music: Music,
  autoPlay: boolean = true,
  set: (status: AnimateStatus) => void = () => {},
) {
  pushToPlaylist(music);
  currentPlayer?.destroy();
  currentPlayer = undefined;

  if (playlist.value[playindex.value] !== music) {
    playindex.value = playlist.value.findIndex(
      (m) => m.bvid === music.bvid && m.cid === music.cid,
    );
  }
  if (randlist.length > 0 && randlist[randindex] !== music) {
    randindex = randlist.findIndex(
      (m) => m.bvid === music.bvid && m.cid === music.cid,
    );
  }

  setPlayState({ index: `${playindex.value}` });

  title.value = music.title;
  cover.value = music.pic;
  progress.value = 0;
  buffered.value = 0;
  duration.value = 0;

  const url = proxyMPD(
    music.bvid,
    music.cid,
    music.title,
    music.pic,
    music.author,
    music.mid,
    music.page,
    qualityPrefer,
  );
  const audio = getAudioPlayer();
  const player = MediaPlayer().create();
  player.updateSettings({
    debug: {
      logLevel: Debug.LOG_LEVEL_WARNING,
    },
    streaming: {
      cacheInitSegments: true,
      enableManifestDurationMismatchFix: true,
      buffer: {
        initialBufferLevel: 0,
        bufferTimeDefault: bufferTime, // 非最高质量时，缓冲N秒
        bufferTimeAtTopQuality: bufferTime, // 最高质量时，缓冲N秒
        bufferTimeAtTopQualityLongForm: bufferTime, // 长内容最高质量时，缓冲N秒
        longFormContentDurationThreshold: 600, // 内容超过N秒视为长内容
        bufferToKeep: bufferTime, // 保留播放位置后N秒的数据
        bufferPruningInterval: 10, // 每N秒检查并清理一次缓冲区
      },
      scheduling: {
        scheduleWhilePaused: false, // 暂停时停止继续调度新的分片请求
      },

      manifestRequestTimeout: 3000, // mpd请求超时
      fragmentRequestTimeout: 5000, // 媒体片段请求超时
      abandonLoadTimeout: 8000, // <放弃加载>的超时时间
      retryIntervals: {
        MPD: 1000,
        MediaSegment: 1000,
        InitializationSegment: 1000,
      },

      retryAttempts: {
        MPD: 3,
        MediaSegment: 3,
        InitializationSegment: 3,
      },
    },
  });
  player.on(MediaPlayer.events.ERROR, (e) => {
    setAnimateStatus(set, 'error');
    logger.error(`player error: ${e.error.toString()}`);
    noticeAlert({
      level: 'alert-error',
      message: `${music.title}：播放失败`,
    });
    failedCount++;
    if (failedCount > 3) {
      return;
    }
    if (playlist.value.length === 1 || playmode.value === 'repeat') {
      setTimeout(() => {
        if (isCurrentPlaying(music.bvid, music.cid)) {
          playMusic(music, autoPlay);
        }
      }, 1000);
    } else if (
      isCurrentPlaying(music.bvid, music.cid) &&
      playlist.value.length > 1
    ) {
      playNext();
    }
  });
  player.on(MediaPlayer.events.MANIFEST_LOADED, (e) => {
    failedCount = 0;
    const dash = DashManifestInfoZ.parse(e.data);
    duration.value = dash.mediaPresentationDuration;
    quality.value = getQualityName(
      dash.Period[0]?.AdaptationSet[0]?.Representation[0]?.id ?? '',
    );
  });
  player.on(MediaPlayer.events.BUFFER_LOADED, () => {
    buffered.value = audio.buffered.end(audio.buffered.length - 1);
  });
  player.on(MediaPlayer.events.PLAYBACK_TIME_UPDATED, (e) => {
    progress.value = e.time || 0;
    buffered.value = audio.buffered.end(audio.buffered.length - 1);
    const dur = Math.floor(audio.duration);
    if (duration.value === 0 || dur !== duration.value) {
      duration.value = dur;
    }
  });
  player.on(MediaPlayer.events.PLAYBACK_SEEKING, (e) => {
    progress.value = e.seekTime || 0;
  });
  player.on(MediaPlayer.events.PLAYBACK_PAUSED, () => {
    playing.value = false;
  });
  let started = false;
  player.on(MediaPlayer.events.PLAYBACK_PLAYING, () => {
    playing.value = true;
    if (!started) {
      started = true;
      setAnimateStatus(set, 'success');
    }
  });
  player.on(MediaPlayer.events.PLAYBACK_ENDED, () => {
    if (playlist.value.length === 1 || playmode.value === 'repeat') {
      player.play();
    } else {
      playNext();
    }
  });
  player.on(MediaPlayer.events.STREAM_TEARDOWN_COMPLETE, () => {
    setAnimateStatus(set, '');
  });
  player.initialize(audio, url, autoPlay);
  currentPlayer = player;

  onPlayCallback(music);

  navigator.mediaSession.metadata = new MediaMetadata({
    title: music.title,
    artist: music.author,
    artwork: [{ src: mustHTTPS(thumbnail(music.pic)) }],
  });
  navigator.mediaSession.setActionHandler('play', player.play);
  navigator.mediaSession.setActionHandler('pause', player.pause);
  if (playlist.value.length > 1) {
    navigator.mediaSession.setActionHandler('nexttrack', () => playNext(true));
    navigator.mediaSession.setActionHandler('previoustrack', () =>
      playPrev(true),
    );
  }
}

export function pause() {
  if (playing.value) {
    currentPlayer?.pause();
  }
}

export function play() {
  if (!playing.value) {
    currentPlayer?.play();
  }
}

export function toggle() {
  if (currentPlayer) {
    if (playing.value) {
      currentPlayer.pause();
    } else {
      currentPlayer.play();
    }
  } else if (playlist.value.length > 0) {
    playNow(playlist.value[0]);
  } else {
    playing.value = false;
  }
}

export function seek(dur: number) {
  currentPlayer?.seek(dur);
}

export function formatDuration(dur: number): string {
  const hours = Math.floor(dur / 3600);
  dur = dur % 3600;
  const minutes = Math.floor(dur / 60);
  const seconds = dur % 60;
  const result =
    `${minutes}`.padStart(2, '0') + ':' + `${seconds}`.padStart(2, '0');
  if (hours > 0) {
    return `${hours}`.padStart(2, '0') + ':' + result;
  }
  return result;
}

export function hasPlayer(): boolean {
  if (!currentPlayer) {
    return false;
  }
  if (playlist.value.length === 0) {
    return false;
  }
  return true;
}

export function isCurrentPlaying(bvid: string, cid: bigint = 0n): boolean {
  if (!hasPlayer()) {
    return false;
  }
  const curr = playlist.value[playindex.value];
  if (bvid !== curr.bvid) {
    return false;
  }
  if (cid && cid !== curr.cid) {
    return false;
  }
  return true;
}

export function currentPlayingClass(bvid: string, cid: bigint = 0n): string {
  return isCurrentPlaying(bvid, cid)
    ? 'bg-linear-to-r from-secondary/10 to-base-100'
    : '';
}
