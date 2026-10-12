<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { PlayIcon, PauseIcon, PreviousIcon, NextIcon } from '@hugeicons/core-free-icons';
import { playing, toggle, playPrev, playNext, seekbackward, seekforward } from '../../libs/player';
import { onMounted } from 'vue';

onMounted(() => {
	window.addEventListener('keydown', (e) => {
		if (e.key === ' ' || e.code === 'Space') {
			toggle();
			e.preventDefault();
		} else if (e.key === 'ArrowLeft' || e.code === 'ArrowLeft') {
			if (e.metaKey) {
				playPrev();
			} else {
				seekbackward();
			}
			e.preventDefault();
		} else if (e.key === 'ArrowRight' || e.code === 'ArrowRight') {
			if (e.metaKey) {
				playNext();
			} else {
				seekforward();
			}
			e.preventDefault();
		}
	})
})
</script>

<template>
	<div class="flex flex-row items-center gap-2">
		<button class="btn btn-circle btn-outline btn-sm" @click.stop="() => playPrev()">
			<HugeiconsIcon :icon="PreviousIcon" :size="20" :stroke-width="1" />
		</button>

		<button class="btn btn-circle btn-outline" @click.stop="toggle">
			<HugeiconsIcon :icon="playing ? PauseIcon : PlayIcon" :size="24" :stroke-width="1" />
		</button>

		<button class="btn btn-circle btn-outline btn-sm" @click.stop="() => playNext()">
			<HugeiconsIcon :icon="NextIcon" :size="20" :stroke-width="1" />
		</button>
	</div>
</template>