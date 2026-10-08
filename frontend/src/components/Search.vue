<script setup lang="ts">
import { computed, onActivated, onDeactivated, ref } from 'vue';
import MainLayout from './MainLayout.vue';
import { setOnScroll } from '../libs/scroll.ts';
import { searchResults, nextPage, parseDuration, hasMore } from '../libs/search.ts';
import { currentKeyword, totalResultCount } from '../libs/search.ts';
import { type Music } from '../libs/player.ts';
import Model from './Model.vue';
import MusicLine from './MusicLine.vue';

onActivated(() => {
	setOnScroll((event: Event) => {
		const target = event.target as HTMLElement;
		const diff = target.scrollHeight - target.clientHeight - target.scrollTop;
		const end = Math.abs(diff) < 100;
		if (end) {
			nextPage();
		}
	})
});

onDeactivated(() => setOnScroll(() => {}));

const multiParts = ref<Music[]>([]);
const showChosenDialog = ref(false);
const musics = computed<Music[]>(() => searchResults.value.map<Music>((item) => ({
	bvid: item.bvid,
	cid: 0n,
	title: item.title,
	desc: item.description,
	duration: parseDuration(item.duration),
	author: item.author,
	mid: `${item.mid}`,
	pic: item.pic,
	play: item.play,
	favorites: item.favorites,
	danmaku: item.danmaku,
	like: item.like,
	pubdate: item.senddate,
})));

</script>

<template>
<div>
	<Model v-model:musics="multiParts" v-model:show="showChosenDialog" />

	<MainLayout v-show="currentKeyword" :title="'搜索：' + currentKeyword" :scroll-to-top="true">
		<ul class="list">
			<li v-if="totalResultCount > 0" class="mb-2 mt-2 text-xs opacity-60 tracking-wide">
				<span>搜索结果</span>
				<span class="ml-2 badge badge-soft badge-secondary badge-xs">
					{{ searchResults.length }}/{{ totalResultCount }}
				</span>
			</li>

			<MusicLine
				v-for="music in musics"
				:key="music.bvid"
				v-model:multi-musics="multiParts"
				v-model:show-model="showChosenDialog"
				:music
				:show-add="true"
			/>
		</ul>

		<div v-if="hasMore" class="flex w-full justify-center mt-8">
			<span class="loading loading-spinner loading-lg"></span>
		</div>
		<span v-else class="flex w-full justify-center mt-8 text-xs opacity-60">——————做人留一线——————</span>
	</MainLayout>
</div>
</template>
