<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { PauseIcon, PlayIcon, Refresh01Icon } from '@hugeicons/core-free-icons';
import { computed, nextTick, onActivated, onMounted, ref } from 'vue';
import MainLayout from '../MainLayout.vue';
import { loadHotMusics, type HotMusic } from '../../libs/recommend.ts';
import { logger, proxyImage } from '../../libs/backend.ts';
import { formatCount, getSongView, type SongView } from '../../libs/search.ts';
import { isCurrentPlaying, pause, play, playing, playNow } from '../../libs/player.ts';
import Empty from '../Empty.vue';
import { setAnimateStatus, type AnimateStatus } from '../../libs/animate.ts';
import LoadingAnimate from '../LoadingAnimate.vue';
import { noticeAlert } from '../../libs/alerts.ts';

type CompHotMusic = HotMusic & {
	animate_status?: AnimateStatus;
}

const playMusic = async (music: CompHotMusic) => {
	const set = (status: AnimateStatus) => music.animate_status = status;
	setAnimateStatus(set, 'loading');
	let dash: SongView;
	try {
		dash = await getSongView(music.bvid);
	} catch (error) {
		setAnimateStatus(set, 'error');
		logger.error(`get song view ${music.bvid}: ${error}`);
		return;
	}
	const pages = dash.pages.filter((p) => `${p.cid}` === music.cid);
	if (pages.length === 0) {
		// Never
		setAnimateStatus(set, 'error');
		return;
	};
	playNow({
		bvid: music.bvid,
		cid: pages[0].cid,
		title: music.music_title,
		desc: dash.desc,
		duration: pages[0].duration,
		author: music.author,
		mid: `${dash.owner.mid}`,
		pic: music.cover,
		page: pages[0].page,
		play: dash.stat.view,
		favorites: dash.stat.favorite,
		danmaku: dash.stat.danmaku,
		like: dash.stat.like,
		pubdate: dash.pubdate,
	})
	setAnimateStatus(set, 'success');
}

const hotMusics = ref<CompHotMusic[]>([]);
const pages = ref<CompHotMusic[][][]>([]);
let pagesize = 3;
const totalPagenum = computed(() => Math.floor((hotMusics.value.length+pagesize*2-1) / (pagesize*2)));
const pagenum = ref(1);

const container = ref();
const elems = ref();
let realGap = 0;
let scroll = 0;
let lastWidth = 0;

const onResize = async () => {
	if (!container.value) return;
	if (!hotMusics.value) return;

	const totalWidth = (container.value as HTMLElement).clientWidth;
	if (totalWidth === lastWidth) {
		return;
	}
	lastWidth = totalWidth;

    logger.debug(`new music container width: ${totalWidth}`);
    const elemWidth = 256; // (elems.value[0] as HTMLElement).offsetWidth;
    logger.debug(`new music container elems: ${elemWidth}`);

	const minGap = 32;
	pagesize = Math.floor((totalWidth - elemWidth) / (elemWidth + minGap)) + 1;
    logger.debug(`new music container pagesize: ${pagesize}`);
    if (pagesize > 1) {
      realGap = (totalWidth - pagesize * elemWidth) / (pagesize - 1);
    } else {
      realGap = 0;
    }
	logger.debug(`new music container gap: ${realGap}`);
    container.value.style = `gap: ${realGap}px`;
    scroll = totalWidth + realGap;
    logger.debug(`new music container scroll: ${scroll}`);

	pagenum.value = 1;
	pages.value = [];
	for (let i = 0; i < totalPagenum.value; i++) {
		pages.value.push([
			hotMusics.value.slice(i*2 * pagesize, (i*2+1) * pagesize),
			hotMusics.value.slice((i*2+1) * pagesize, (i*2+2) * pagesize),
		]);
	}

	await nextTick();
	for (const elem of elems.value) {
		elem.style = `gap: ${realGap}px`;
	}
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
	const canSet = hotMusics.value.length === 0;
	const set = (status: AnimateStatus) => {
		if (canSet) {
			refreshStatus.value = status;
		}
	};
	setAnimateStatus(set, 'loading');
	try {
		hotMusics.value = await loadHotMusics();
		await nextTick();
		onResize();
		setTimeout(refresh, 10 * 3600_000);
		setAnimateStatus(set, 'success');
	} catch (error) {
		if (hotMusics.value.length === 0) {
			noticeAlert({level: 'alert-error', message: `拉取热歌精选失败：${error}`});
		}
		logger.error(`load hot musics: ${error}`);
		if (hotMusics.value.length > 0) {
			setTimeout(refresh, 10 * 3600_000);
		}
		setAnimateStatus(set, 'error');
	} finally {
		isLoading = false;
		setTimeout(refresh, 10 * 3600_000);
	}
};

onMounted(() => {
	window.addEventListener('resize', onResize);
	refresh();
});

onActivated(onResize);
</script>

<template>
	<MainLayout title="热歌精选">
		<template #header>
			<LoadingAnimate
				:status="refreshStatus"
				out-size="size-8"
				:icon-size="24"
			>
				<div v-if="hotMusics.length > 0" class="join flex items-center border-0">
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
				v-if="hotMusics.length > 0"
				class="w-full flex flex-nowrap relative duration-300"
			>
				<div
					v-for="page in pages"
					class="w-full flex flex-col gap-8"
				>
					<div
						v-for="musics in page"
						class="w-full flex justify-between"
						ref="elems"
					>
						<div
							v-for="music in musics"
							class="card bg-base-100 shadow-0 w-64 shrink-0 border border-base-300"
						>
							<figure>
								<img
									:src="proxyImage(music.cover, 0)"
									class="w-full h-36 object-cover"
								/>
							</figure>
							<div class="card-body p-2 justify-between">
								<h2 class="card-title line-clamp-1">{{ music.music_title }}</h2>
								<div class="card-actions flex flex-row flex-nowrap justify-between items-center gap-2">
									<div class="grow flex flex-col gap-1 text-xs opacity-60">
										<span class="line-clamp-1">{{ music.author }} · {{ music.album }}</span>
										<span>
											<HugeiconsIcon class="inline me-0.5" :size="12" :icon="PlayIcon" :stroke-width="1"  />
											{{ formatCount(music.total_vv) }}
										</span>
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
				</div>
			</div>
			<Empty v-else />
		</div>
	</MainLayout>
</template>