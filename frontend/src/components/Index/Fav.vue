<script setup lang="ts">
  import { HugeiconsIcon } from "@hugeicons/vue";
  import {
    Refresh01Icon,
    Folder01Icon,
    Folder02Icon,
    PlayIcon,
    CirclePlusIcon,
  } from "@hugeicons/core-free-icons";
  import { onMounted, ref } from "vue";
  import {
    getFavorites,
    getFavResources,
    MediaAttr,
    type FavMedia,
    type FavoriteFolder,
  } from "../../libs/fav.ts";
  import { isLogined } from "../../libs/backend.ts";
  import MainLayout from "../MainLayout.vue";
  import {
    playNow,
    pushMultiToPlaylist,
    type Music,
  } from "../../libs/player.ts";
  import { type AnimateStatus, setAnimateStatus } from "../../libs/animate.ts";
  import Model from "../Model.vue";
  import { registerOnLoginStatusChanged } from "../../libs/login.ts";
  import { noticeAlert } from "../../libs/alerts.ts";
  import LoadingAnimate from "../LoadingAnimate.vue";
  import Empty from "../Empty.vue";
  import MusicLine from "../MusicLine.vue";
  import { playMusicByBvid } from "../../libs/common.ts";

  type CompFavMedia = Music & {
    attr: number;
    totalPage: number;
  };

  type FavFolder = {
    folder: FavoriteFolder;
    medias: CompFavMedia[];
    has_more: boolean;
    spread: boolean;
    push_status?: AnimateStatus;
    play_status?: AnimateStatus;
  };

  const favFolders = ref<FavFolder[]>([]);
  const refreshStatus = ref("");

  const refresh = async () => {
    refreshStatus.value = "loading";
    try {
      if (!(await isLogined())) {
        favFolders.value = [];
        return;
      }

      const favs = await getFavorites();
      const ori = favFolders.value.map((ff) => ff.folder);
      if (favs.length === ori.length) {
        let equal = favs
          .map((fav, i) => fav === ori[i])
          .reduce((prev, curr) => prev && curr, true);
        if (equal) {
          refreshStatus.value = "success";
          return;
        }
      }
      const folders = favs.map((fav) => ({
        folder: fav,
        medias: [],
        has_more: true,
        spread: false,
        push_status: "" as const,
        play_status: "" as const,
      }));
      favFolders.value = folders;
      refreshStatus.value = "success";
    } catch (error) {
      refreshStatus.value = "error";
      noticeAlert({
        level: "alert-error",
        message: `加载“我创建的收藏夹”失败：${error}`,
      });
    } finally {
      setTimeout(() => (refreshStatus.value = ""), 1000);
    }
  };

  onMounted(() => {
    refresh();
    registerOnLoginStatusChanged("index-fav", refresh);
  });

  const foldStatus = ref("");
  const foldAll = () => {
    foldStatus.value = "loading";
    favFolders.value.forEach((fav) => (fav.spread = false));
    foldStatus.value = "success";
    setTimeout(() => (foldStatus.value = ""), 1000);
  };

  const loadFav = async (fav: FavFolder) => {
    let pn = 0;
    while (fav.has_more) {
      pn++;
      const result = await getFavResources(fav.folder.id, pn);
      fav.has_more = result.has_more;

      const medias: CompFavMedia[] = [];
      for (const media of result.medias) {
        doOneMedia(media, (music) => {
          medias.push({ ...music, attr: media.attr, totalPage: media.page });
        });
      }
      fav.medias = [...fav.medias, ...medias];
    }
  };

  const loadAndSpread = async (fav: FavFolder) => {
    if (fav.spread) {
      fav.spread = false;
      return;
    }

    fav.spread = true;
    try {
      await loadFav(fav);
    } catch (error) {
      fav.spread = false;
      noticeAlert({
        level: "alert-error",
        message: `获取收藏夹《${fav.folder.title}》数据失败：${error}`,
      });
      console.log(`load fav ${fav.folder.title}: ${error}`);
    }
  };

  const multiParts = ref<Music[]>([]);
  const showChosenDialog = ref<boolean>(false);

  const doOneMedia = (media: FavMedia, action: (music: Music) => void) => {
    action({
      bvid: media.bvid,
      cid: media.page === 1 ? media.ugc?.first_cid || 0n : 0n,
      title: media.title,
      desc: media.intro,
      duration: media.duration,
      author: media.upper.name,
      mid: `${media.upper.mid}`,
      pic: media.cover,
      play: media.cnt_info.play,
      favorites: media.cnt_info.collect,
      danmaku: media.cnt_info.danmaku,
      pubdate: media.pubtime,
    });
  };

  const pushMultiAndPlay = async (fav: FavFolder, play: boolean = false) => {
    let ok = "已加入播放列表";
    if (play) ok = "已开始播放";
    let failed = "添加失败";
    if (play) failed = "播放失败";

    let set: (status: AnimateStatus) => void;
    if (play) {
      set = (status) => (fav.play_status = status);
    } else {
      set = (status) => (fav.push_status = status);
    }
    setAnimateStatus(set, "loading");

    try {
      await _pushMultiAndPlay(fav, play);
      setAnimateStatus(set, "success");
      noticeAlert({ level: "alert-info", message: ok });
    } catch (error) {
      setAnimateStatus(set, "error");
      noticeAlert({ level: "alert-error", message: `${failed}：${error}` });
      console.log(`push all musics of fav ${fav.folder.title}: ${error}`);
    }
  };

  const _pushMultiAndPlay = async (fav: FavFolder, play: boolean = false) => {
    await loadFav(fav);
    let all: Music[] = [];
    for (const media of fav.medias) {
      if (media.attr !== MediaAttr.Normal) {
        continue;
      }
      if (media.cid !== 0n) {
        all.push(media);
        continue;
      }

      await playMusicByBvid(
        media.bvid,
        (music) => all.push(music),
        (musics) => (all = [...all, ...musics]),
      );
    }
    pushMultiToPlaylist(all);
    if (play && all.length > 0) {
      playNow(all[0]);
    }
  };
</script>

<template>
  <Model v-model:musics="multiParts" v-model:show="showChosenDialog" />

  <MainLayout title="我创建的收藏夹">
    <template #header>
      <div class="flex items-center gap-4">
        <LoadingAnimate :status="foldStatus" out-size="size-8" :icon-size="24">
          <button
            class="btn btn-sm btn-circle btn-outline tooltip"
            data-tip="收起所有"
            @click.stop="foldAll"
          >
            <HugeiconsIcon :icon="Folder01Icon" :size="18" />
          </button>
        </LoadingAnimate>

        <LoadingAnimate
          :status="refreshStatus"
          out-size="size-8"
          :icon-size="24"
        >
          <button
            class="btn btn-sm btn-circle btn-outline tooltip"
            data-tip="刷新"
            @click.stop="refresh"
          >
            <HugeiconsIcon :icon="Refresh01Icon" :size="18" />
          </button>
        </LoadingAnimate>
      </div>
    </template>

    <div v-if="favFolders.length > 0" class="flex flex-col gap-4 my-6">
      <div
        v-for="fav in favFolders"
        :key="fav.folder.id"
        class="hover:bg-base-200 flex flex-col rounded-box shadow-md border border-base-300"
        :class="
          fav.spread
            ? 'hover:bg-transparent bg-linear-to-r from-primary/10 to-base-100'
            : ''
        "
      >
        <div class="flex items-center p-4" @click.stop="loadAndSpread(fav)">
          <div class="grow flex gap-2 items-center">
            <HugeiconsIcon
              v-if="fav.spread"
              :icon="Folder02Icon"
              :size="18"
              :stroke-width="1"
            />
            <HugeiconsIcon
              v-else
              :icon="Folder01Icon"
              :size="18"
              :stroke-width="1"
            />
            <h2 class="line-clamp-1">{{ fav.folder.title }}</h2>
            <span class="badge badge-xs">{{ fav.folder.media_count }}</span>
          </div>
          <div class="flex gap-4">
            <LoadingAnimate
              :status="fav.push_status"
              out-size="size-6"
              :icon-size="24"
            >
              <div
                class="tooltip"
                data-tip="全部加入播放列表"
                @click.stop="pushMultiAndPlay(fav)"
              >
                <HugeiconsIcon
                  class="hover:text-secondary"
                  :icon="CirclePlusIcon"
                  :stroke-width="1"
                />
              </div>
            </LoadingAnimate>
            <LoadingAnimate
              :status="fav.play_status"
              out-size="size-6"
              :icon-size="24"
            >
              <div
                class="tooltip"
                data-tip="播放全部"
                @click.stop="pushMultiAndPlay(fav, true)"
              >
                <HugeiconsIcon
                  class="hover:text-secondary"
                  :icon="PlayIcon"
                  :stroke-width="1"
                />
              </div>
            </LoadingAnimate>
          </div>
        </div>

        <ul v-if="fav.spread" class="list bg-base-100 rounded-box shadow-md">
          <MusicLine
            v-model:multi-musics="multiParts"
            v-model:show-model="showChosenDialog"
            v-for="media in fav.medias"
            :key="media.bvid"
            :music="media"
            :attr="media.attr"
            :show-add="true"
          />
        </ul>
      </div>
    </div>
    <Empty v-else />
  </MainLayout>
</template>
