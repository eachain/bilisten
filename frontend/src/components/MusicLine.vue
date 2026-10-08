<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import {
	PlayIcon,
	PauseIcon,
	Delete02Icon,
	CirclePlusIcon,
	HeadphonesIcon,
	AppleReminderIcon,
	Bookmark02Icon,
	ThumbsUpIcon,
	UserAccountIcon,
	Calendar03Icon,
	ExternalLinkIcon,
} from '@hugeicons/core-free-icons';
import { type CSSProperties, ref } from 'vue';
import {
	formatDuration,
	isCurrentPlaying,
	pause,
	play,
	playNow,
	playing,
	popFromPlaylist,
	pushToPlaylist,
	type Music
} from '../libs/player.ts';
import { proxyImage, userSpaceURL, videoURL } from '../libs/backend.ts';
import { formatCount, formatSendDate } from '../libs/search.ts';
import { setAnimateStatus, type AnimateStatus } from '../libs/animate.ts';
import { playMusicByBvid } from '../libs/common.ts';
import { MediaAttr } from '../libs/fav.ts';
import { openURL } from '../libs/wails.ts';
import LoadingAnimate from './LoadingAnimate.vue';

const multiParts = defineModel<Music[]>('multi-musics');
const showChosenDialog = defineModel<boolean>('show-model');

const { music, attr, stylex, showAdd, showDel } = defineProps<{
	music: Music,
	attr?: number, // see fav.MediaAttr
	stylex?: CSSProperties,
	showAdd?: boolean, // show add to playlist icon
	showDel?: boolean, // show del from playlist icon
}>();

const playStatus = ref<AnimateStatus>('');
const pushStatus = ref<AnimateStatus>('');

const setPlayStatus = (status: AnimateStatus) => playStatus.value = status;
const setPushStatus = (status: AnimateStatus) => pushStatus.value = status;

const playMusic = async (music: Music) => {
	if ((attr ?? MediaAttr.Normal) !== MediaAttr.Normal) {
        return;
    }

	setAnimateStatus(setPlayStatus, 'loading');
	if (music.cid) {
		playNow(music, setPlayStatus);
		return;
	}

	try {
		await playMusicByBvid(
			music.bvid,
			(music) => playNow(music, setPlayStatus),
			(musics) => {
				multiParts.value = musics;
				showChosenDialog.value = true;
				setAnimateStatus(setPlayStatus, '');
			},
		)
	} catch (error) {
		setAnimateStatus(setPlayStatus, 'error');
		console.log(`get song view of ${music.bvid}: ${error}`);
	}
}

const pushMusic = async (music: Music) => {
	if ((attr ?? MediaAttr.Normal) !== MediaAttr.Normal) {
        return;
    }

	setAnimateStatus(setPushStatus, 'loading');
	if (music.cid) {
		pushToPlaylist(music);
		setAnimateStatus(setPushStatus, 'success');
		return;
	}

	try {
		await playMusicByBvid(
			music.bvid,
			(music) => {
				pushToPlaylist(music);
				setAnimateStatus(setPushStatus, 'success');
			},
			(musics) => {
				multiParts.value = musics;
				showChosenDialog.value = true;
				setAnimateStatus(setPushStatus, '');
			},
		)
	} catch (error) {
		setAnimateStatus(setPushStatus, 'error');
		console.log(`get song view of ${music.bvid}: ${error}`);
	}
};

</script>

<template>
	<li
		class="list-row flex w-full hover:bg-base-200"
		:class="isCurrentPlaying(music.bvid, music.cid) ? 'bg-linear-to-r from-secondary/10 to-base-100' : ''"
		:id="'playlist-' + music.bvid + '-' + music.cid"
		:style="stylex"
		@dblclick.stop="playNow(music)"
	>
		<div class="w-full flex flex-row items-center gap-4">
			<div class="flex gap-2 grow overflow-hidden">
				<div>
					<img class="size-10 min-w-10 rounded-box object-cover" :alt="music.title" :src="proxyImage(music.pic)"/>
				</div>
				<div
					class="grow flex flex-col justify-between overflow-hidden"
					:class="(attr ?? MediaAttr.Normal) !== MediaAttr.Normal ? 'opacity-60' : ''"
				>
					<div class="line-clamp-1" v-html="music.title"></div>
					<div class="max-w-fit text-xs opacity-60 flex gap-3 tabular-nums truncate">
						<span class="flex items-center">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="HeadphonesIcon" :stroke-width="1"  />
							{{ formatDuration(music.duration) }}
						</span>
						<span v-if="music.play" class="flex items-center">
							<HugeiconsIcon class="opacity-50 inline me-px" :size="12" :icon="PlayIcon" :stroke-width="1"  />
							{{ formatCount(music.play) }}
						</span>
						<span v-if="music.danmaku" class="flex items-center">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="AppleReminderIcon" :stroke-width="1"  />
							{{ formatCount(music.danmaku) }}
						</span>
						<span v-if="music.favorites" class="flex items-center">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="Bookmark02Icon" :stroke-width="1"  />
							{{ formatCount(music.favorites) }}
						</span>
						<span v-if="music.like" class="flex items-center">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="ThumbsUpIcon" :stroke-width="1"  />
							{{ formatCount(music.like) }}
						</span>
						<span class="flex items-center hover:text-secondary cursor-pointer" @click="openURL(userSpaceURL(music.mid))">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="UserAccountIcon" :stroke-width="1"  />
							{{ music.author }}
						</span>
						<span v-if="music.pubdate" class="flex items-center">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="Calendar03Icon" :stroke-width="1"  />
							{{ formatSendDate(music.pubdate) }}
						</span>
						<span class="flex items-center hover:text-secondary cursor-pointer" @click="openURL(videoURL(music.bvid, music.page))">
							<HugeiconsIcon class="opacity-50 inline me-0.5" :size="12" :icon="ExternalLinkIcon" :stroke-width="1"  />
							原视频
						</span>
					</div>
				</div>
			</div>

			<div v-if="(attr ?? MediaAttr.Normal) === MediaAttr.Normal" class="flex gap-4">
				<LoadingAnimate v-if="showAdd" :status="pushStatus" out-size="size-6" :icon-size="24">
					<div class="tooltip" data-tip="加入播放列表" @click.stop="pushMusic(music)">
						<HugeiconsIcon class="hover:text-secondary" :icon="CirclePlusIcon" :stroke-width="1" />
					</div>
				</LoadingAnimate>

				<div v-if="showDel" class="tooltip" data-tip="删除" @click.stop="popFromPlaylist(music)">
					<HugeiconsIcon class="hover:text-error" :icon="Delete02Icon" :stroke-width="1" />
				</div>

				<LoadingAnimate :status="playStatus" out-size="size-6" :icon-size="24">
					<div v-if="playing && isCurrentPlaying(music.bvid, music.cid)" class="swap-on tooltip" data-tip="暂停" @click.stop="pause">
						<HugeiconsIcon class="hover:text-secondary" :icon="PauseIcon" :stroke-width="1" />
					</div>
					<div v-else-if="isCurrentPlaying(music.bvid, music.cid)" class="swap-on tooltip" data-tip="继续" @click.stop="play">
						<HugeiconsIcon class="hover:text-secondary" :icon="PlayIcon" :stroke-width="1" />
					</div>
					<div v-else class="swap-off tooltip" data-tip="播放" @click.stop="playMusic(music)">
						<HugeiconsIcon class="hover:text-secondary" :icon="PlayIcon" :stroke-width="1" />
					</div>
				</LoadingAnimate>
			</div>
		</div>
	</li>
</template>

<style lang="css">
.keyword {
    color: var(--color-secondary);
    display: inline;
    font-style: normal;
}
</style>
