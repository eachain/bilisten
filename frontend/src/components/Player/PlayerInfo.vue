<script setup lang="ts">
import { computed } from 'vue';
import musicImage from '../../assets/music.png';
import { title, cover, duration, progress, formatDuration } from '../../libs/player.ts';
import Slider from './Slider.vue';
import { proxyImage } from '../../libs/backend.ts';

const musicCover = computed(() => {
	if (cover.value) {
		return proxyImage(cover.value);
	}
	return musicImage;
});
</script>

<template>
	<div class="flex flex-row gap-2 w-full">
		<div class="avatar">
			<div class="w-12 rounded-xl">
				<img alt="" :src="musicCover" />
			</div>
		</div>

		<div class="grow py-0.5 flex flex-col justify-between">
			<span class="line-clamp-1">{{ title }}</span>
			<div class="flex flex-row gap-2 items-center">
				<Slider />
				<span class="text-xs font-light opacity-75 tabular-nums">
					{{formatDuration(Math.floor(progress))}}/{{ formatDuration(duration) }}
				</span>
			</div>
		</div>
	</div>
</template>