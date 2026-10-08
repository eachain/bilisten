<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import {
	PlayIcon,
	Calendar03Icon,
	PauseIcon,
	Refresh01Icon,
} from '@hugeicons/core-free-icons';
import { computed, nextTick, onActivated, onMounted, ref } from 'vue';
import MainLayout from '../MainLayout.vue';
import { loadNewMusics, type NewMusicInfo } from '../../libs/recommend.ts';
import { proxyImage } from '../../libs/backend.ts';
import { formatCount, getSongView, type SongView } from '../../libs/search.ts';
import { isCurrentPlaying, type Music, pause, play, playing, playNow, pushMultiToPlaylist } from '../../libs/player.ts';
import { noticeAlert } from '../../libs/alerts.ts';
import Empty from '../Empty.vue';
import { type AnimateStatus, setAnimateStatus } from '../../libs/animate.ts';
import LoadingAnimate from '../LoadingAnimate.vue';

type CompNewMusicInfo = NewMusicInfo & {
	animate_status?: AnimateStatus;
}

const playMusic = async (music: CompNewMusicInfo) => {
	const set = (status: AnimateStatus) => music.animate_status = status;
	setAnimateStatus(set, 'loading');
	let dash: SongView;
	try {
		dash = await getSongView(music.bvid);
	} catch (error) {
		setAnimateStatus(set, 'error');
		console.log(`get song view of ${music.bvid}: ${error}`);
		return
	}
	if (dash.pages.length === 0) {
		// Never
		setAnimateStatus(set, 'error');
		return;
	};
	if (dash.pages.length === 1) {
		playNow({
			bvid: music.bvid,
			cid: dash.cid,
			title: music.music_title,
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
		})
		setAnimateStatus(set, 'success');
		return;
	}
	
	const multiParts: Music[] = [];
	for (const page of dash.pages) {
		multiParts.push({
			bvid: dash.bvid,
			cid: page.cid,
			title: page.part,
			desc: dash.desc,
			duration: page.duration,
			author: music.author,
			mid: `${dash.owner.mid}`,
			page: page.page,
			pic: music.cover,
			pubdate: dash.pubdate,
		})
	}
	pushMultiToPlaylist(multiParts);
	playNow(multiParts[0]);
	setAnimateStatus(set, 'success');
	noticeAlert({level: 'alert-info', message: '已开始播放'});
}

const newMusics = ref<CompNewMusicInfo[]>([]);
let pagesize = 4;
const totalPagenum = computed(() => Math.floor((newMusics.value.length+pagesize-1) / pagesize));
const pagenum = ref(1);

const container = ref();
const elems = ref();
let realGap = 0;
let scroll = 0;
let lastWidth = 0;

const onResize = () => {
	if (!container.value) {
		return;
	}

	const totalWidth = (container.value as HTMLElement).clientWidth;
	if (totalWidth === lastWidth) {
		return;
	}
	lastWidth = totalWidth;

    console.log("new music container width:", totalWidth);
    const elemWidth = (elems.value[0] as HTMLElement).offsetWidth;
    console.log("new music container elems:", elemWidth);

	const minGap = 32;
	pagesize = Math.floor((totalWidth - elemWidth) / (elemWidth + minGap)) + 1;
    console.log("new music container pagesize:", pagesize);
    if (pagesize > 1) {
      realGap = (totalWidth - pagesize * elemWidth) / (pagesize - 1);
    } else {
      realGap = 0;
    }
	console.log("new music container gap:", realGap);
    container.value.style = `gap: ${realGap}px`;
	pagenum.value = 1;
    scroll = totalWidth + realGap;
    console.log("new music container scroll:", scroll);
}

const prevPage = () => {
	pagenum.value--;
	const px = scroll * (pagenum.value - 1);
	container.value.style = `gap: ${realGap}px; transform: translateX(-${px}px)`;
};

const nextPage = () => {
	pagenum.value++;
	const px = scroll * (pagenum.value - 1);
	container.value.style = `gap: ${realGap}px; transform: translateX(-${px}px)`;
};

const refreshStatus = ref<AnimateStatus>('');
let isLoading = false;
const refresh = async () => {
	if (isLoading) {
		return;
	}
	isLoading = true;
	const canSet = newMusics.value.length === 0;
	const set = (status: AnimateStatus) => {
		if (canSet) {
			refreshStatus.value = status;
		}
	};
	setAnimateStatus(set, 'loading');
	try {
		newMusics.value = await loadNewMusics();
		await nextTick();
		onResize();
		setTimeout(refresh, 10 * 3600_000);
		setAnimateStatus(set, 'success');
	} catch (error) {
		if (newMusics.value.length === 0) {
			noticeAlert({level: 'alert-error', message: `拉取新歌速递失败：${error}`});
		}
		console.log('load new musics:', error);
		if (newMusics.value.length > 0) {
			setTimeout(refresh, 10 * 3600_000);
		}
		setAnimateStatus(set, 'error');
	} finally {
		isLoading = false;
		setTimeout(refresh, 10 * 3600_000);
	}
};

onMounted(async () => {
	window.addEventListener('resize', onResize);
	refresh();
});

onActivated(onResize);
</script>

<template>
	<MainLayout title="新歌速递">
		<template #header>
			<LoadingAnimate
				:status="refreshStatus"
				out-size="size-8"
				:icon-size="24"
			>
				<div v-if="newMusics.length > 0" class="join flex items-center border-0">
					<button
						class="join-item btn btn-sm"
						:disabled="pagenum <= 1"
						@click="prevPage"
					>«</button>
					<span class="join-item border-0 px-2 tabular-nums">{{ pagenum }} / {{ totalPagenum }}</span>
					<button
						class="join-item btn btn-sm"
						:disabled="pagenum >= totalPagenum"
						@click="nextPage"
					>»</button>
				</div>
				<button
					v-else
					class="btn btn-sm btn-circle btn-outline tooltip"
					data-tip="刷新"
					@click.stop="refresh"
				>
					<HugeiconsIcon :icon="Refresh01Icon" :size="18" />
				</button>
			</LoadingAnimate>
		</template>

		<div class="w-full overflow-hidden">
			<div
				ref="container"
				v-if="newMusics.length > 0"
				class="w-full flex flex-nowrap relative duration-300"
			>
				<div
					v-for="music in newMusics"
					class="card shadow-0 w-64 shrink-0 border border-base-300"
					ref="elems"
				>
					<figure>
						<img
							:src="proxyImage(music.cover, 0)"
							class="w-full h-36 object-cover"
						/>
					</figure>
					<div class="card-body grow-0 p-2 gap-2">
						<h2 class="card-title line-clamp-1">{{ music.music_title }}</h2>
						<div class="card-actions flex flex-row flex-nowrap justify-between items-center gap-2">
							<div class="grow flex flex-col gap-1 text-xs opacity-60">
								<span class="line-clamp-1">{{ music.author }} · {{ music.album }}</span>
								<div class="flex gap-2">
									<span>
										<HugeiconsIcon class="inline me-0.5" :size="12" :icon="PlayIcon" :stroke-width="1"  />
										{{ formatCount(music.total_vv) }}
									</span>
									<span class="flex items-center">
										<HugeiconsIcon class="inline me-0.5" :size="12" :icon="Calendar03Icon" :stroke-width="1"  />
										{{ music.publish_time }}
									</span>
								</div>
							</div>

							<LoadingAnimate :status="music.animate_status" out-size="size-6" :icon-size="24">
								<div v-if="playing && isCurrentPlaying(music.bvid)" class="tooltip" data-tip="暂停" @click.stop="pause">
									<HugeiconsIcon class="hover:text-secondary" :icon="PauseIcon" :size="24" :stroke-width="1" />
								</div>
								<div v-else-if="isCurrentPlaying(music.bvid)" class="tooltip" data-tip="继续" @click.stop="play">
									<HugeiconsIcon class="hover:text-secondary" :icon="PlayIcon" :size="24" :stroke-width="1" />
								</div>
								<div v-else class="tooltip" data-tip="播放" @click.stop="playMusic(music)">
									<HugeiconsIcon class="hover:text-secondary" :icon="PlayIcon" :size="24" :stroke-width="1" />
								</div>
							</LoadingAnimate>
						</div>
					</div>
				</div>
			</div>
			<Empty v-else />
		</div>
	</MainLayout>
</template>
