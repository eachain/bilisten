<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { ShuffleIcon, RepeatIcon, RepeatOne01Icon } from '@hugeicons/core-free-icons';
// import { VolumeMute01Icon, VolumeMute02Icon, VolumeLowIcon, VolumeHighIcon } from '@hugeicons/core-free-icons';
import { Playlist03Icon } from '@hugeicons/core-free-icons';

import { computed } from 'vue';

import { chmod, playmode, quality } from '../../libs/player.ts';
import { getQualityName } from '../../libs/backend.ts';

const modeSetting = new Map()
	.set('list', {
		name: "列表循环",
		icon: RepeatIcon,
	})
	.set('repeat', {
		name: "单曲循环",
		icon: RepeatOne01Icon,
	})
	.set('shuffle', {
		name: "随机播放",
		icon: ShuffleIcon,
	});

const mode = computed(() => modeSetting.get(playmode.value));

const showPlaylist = () => window.location.hash = '#/playlist';


</script>

<template>
	<div class="flex flex-col flex-1">
		<div class="flex flex-row gap-8 items-center">
			<span class="badge badge-outline tooltip text-xs hover:text-secondary" data-tip="音质">
				{{ getQualityName(quality) }}
			</span>
			<div @click="chmod" class="tooltip" :data-tip="mode.name">
				<HugeiconsIcon class="hover:text-secondary" :icon="mode.icon" :stroke-width="1" />
			</div>

			<!-- 
			<div class="dropdown dropdown-top dropdown-center">
				<div tabindex="0" role="button" class="m-1">
					<HugeiconsIcon class="hover:text-secondary" :icon="VolumeHighIcon" :stroke-width="1" />
				</div>
				<div tabindex="-1" class="dropdown-content bg-base-100 rounded-box z-1 p-2 shadow-sm -rotate-90 mb-42">
				</div>
			</div>

			<span>
				<HugeiconsIcon :icon="VolumeMute02Icon" :stroke-width="1" />
			</span>
			<span>
				<HugeiconsIcon :icon="VolumeMute01Icon" :stroke-width="1" />
			</span>
			<span>
				<HugeiconsIcon :icon="VolumeLowIcon" :stroke-width="1" />
			</span>
			<span>
				<HugeiconsIcon :icon="VolumeHighIcon" :stroke-width="1" />
			</span> -->

			<div @click="showPlaylist" class="tooltip" data-tip="播放列表">
				<HugeiconsIcon class="hover:text-secondary" :icon="Playlist03Icon" :stroke-width="1" />
			</div>
		</div>
	</div>
</template>