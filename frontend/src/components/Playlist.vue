<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import {
	Delete02Icon,
	CrosshairIcon,
} from '@hugeicons/core-free-icons';
import {useRecycleScroller} from 'vue-virtual-scroller';
import MainLayout from './MainLayout.vue'
import {
	clearCurrentPlaylist,
	hasPlayer,
	playindex,
	playlist,
	setOnPlay,
	type Music
} from '../libs/player.ts';
import { ref, onActivated, onDeactivated, nextTick, computed } from 'vue';
import MusicLine from './MusicLine.vue';
import { setScrollToTop } from '../libs/scroll.ts';

const multiParts = ref<Music[]>([]);
const showChosenDialog = ref(false);

const options = computed(() => ({
	items: playlist.value,
	keyField: (item: Music) => `${item.bvid}-${item.cid}`,
	direction: 'vertical' as const,
	itemSize: 72,
	gridItems: undefined,
	itemSecondarySize: undefined,
	minItemSize: null,
	sizeField: 'size',
	typeField: 'type',
	buffer: 200,
	pageMode: false,
	prerender: 0,
	emitUpdate: false,
	updateInterval: 0,
}));

const {
	pool,
	totalSize,
	getViewStyle,
	scrollToPosition,
	getItemOffset,
} = useRecycleScroller(options, () => document.getElementById('app-main')!);

const scrollToCurrent = () => {
	if (playindex.value <= 3) {
		scrollToPosition(0, {smooth: true});
		return;
	}
	scrollToPosition(getItemOffset(playindex.value-3)+105, {smooth: true});
};

const scrollToPlaying = () => {
	if (hasPlayer()) {
		scrollToCurrent();
	}
}

onActivated(async () => {
	setScrollToTop(() => scrollToPosition(0));
	setOnPlay(scrollToCurrent);
	if (!hasPlayer()) return;
	await nextTick();
	scrollToCurrent();
})

onDeactivated(() => {
	setOnPlay(() => {});
});

</script>

<template>
<div>
	<div class="fab">
		<button
			class="btn btn-lg btn-circle btn-outline btn-sm mb-20 mr-2 tooltip tooltip-left"
			data-tip="定位到当前播放"
			@click="scrollToPlaying"
		>
			<HugeiconsIcon :size="18" :icon="CrosshairIcon" />
		</button>
	</div>

	<MainLayout title="播放列表">
		<template #header>
			<div class="flex gap-4 px-4">
				<div class="tooltip tooltip-left btn btn-circle btn-outline btn-sm hover:bg-error" data-tip="清空播放列表" @click="clearCurrentPlaylist">
					<HugeiconsIcon :size="18" :icon="Delete02Icon" />
				</div>
			</div>
		</template>

		<div v-if="playlist.length > 0" class="mb-2 mt-2 text-xs opacity-60 tracking-wide">
			<span>当前播放</span>
			<span v-if="hasPlayer()" class="ml-2 badge badge-soft badge-secondary badge-xs">
				{{ playindex+1 }}/{{ playlist.length }}
			</span>
		</div>

		<ul class="list relative overflow-hidden" :style="{ minHeight: `${totalSize}px` }">
			<MusicLine
				v-for="music in pool"
				:key="music.nr.id"
				v-model:multi-musics="multiParts"
				v-model:show-model="showChosenDialog"
				:music="music.item"
				:stylex="getViewStyle(music)"
				:show-del="true"
			/>
		</ul>
	</MainLayout>
</div>
</template>
