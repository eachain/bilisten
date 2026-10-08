<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import {
	PlayIcon,
	PauseIcon,
	CirclePlusIcon,
} from '@hugeicons/core-free-icons';
import { ref, computed } from 'vue';
import { isCurrentPlaying, pause, play, playing, playNow, pushMultiToPlaylist, type Music } from '../libs/player';
import { noticeAlert } from '../libs/alerts';
import MusicLine from './MusicLine.vue';

const multiParts = defineModel<Music[]>('musics');
const showChosenDialog = defineModel<boolean>('show');
const chosenDialog = computed(() => showChosenDialog.value ? 'modal-open' : '');

const closeChosenDialog = () => {
	showChosenDialog.value = false;
	multiParts.value = [];
}

const pushing = ref(false);

const pushMultiAndPlay = (musics: Music[], play: boolean = false) => {
	pushing.value = true;
	pushMultiToPlaylist(musics);
	pushing.value = false;
	if (play && musics.length > 0) {
		playNow(musics[0]);
		noticeAlert({level: 'alert-info', message: '已开始播放'});
	} else {
		noticeAlert({level: 'alert-info', message: '已加入播放列表'});
	}
	closeChosenDialog();
}

</script>

<template>
	<dialog class="modal" :class="chosenDialog">
		<div class="modal-box w-11/12 max-w-5xl h-3/4 max-h-4/5 p-0">
			<div class="w-full h-full flex flex-col gap-0">
				<div class="bg-base-300 px-8 py-4 flex justify-between shadow-sm">
					<h3 class="text-lg font-bold">请选择</h3>
					<div v-if="pushing" class="loading loading-spinner loading-sm"></div>
					<div v-else class="flex gap-4">
						<div class="tooltip tooltip-bottom" data-tip="全部加入播放列表" @click.stop="pushMultiAndPlay(multiParts ?? [])">
							<HugeiconsIcon class="hover:text-secondary" :icon="CirclePlusIcon" :stroke-width="1" />
						</div>
						<div v-if="playing && isCurrentPlaying(multiParts?.[0]?.bvid ?? '')" class="tooltip tooltip-bottom" data-tip="暂停" @click.stop="pause">
							<HugeiconsIcon class="hover:text-secondary" :icon="PauseIcon" :stroke-width="1" />
						</div>
						<div v-else-if="isCurrentPlaying(multiParts?.[0]?.bvid ?? '')" class="tooltip tooltip-bottom" data-tip="继续" @click.stop="play">
							<HugeiconsIcon class="hover:text-secondary" :icon="PlayIcon" :stroke-width="1" />
						</div>
						<div v-else class="tooltip tooltip-bottom" data-tip="播放全部" @click.stop="pushMultiAndPlay(multiParts ?? [], true)">
							<HugeiconsIcon class="hover:text-secondary" :icon="PlayIcon" :stroke-width="1" />
						</div>
					</div>
				</div>
				<div class="grow overflow-clip overflow-y-auto">
					<ul class="list px-4">
						<MusicLine
							v-for="music in multiParts"
							:key="music.bvid + '-' + music.cid"
							v-model:multi-musics="multiParts"
							v-model:show-model="showChosenDialog"
							:music
							:show-add="true"
						/>
					</ul>
				</div>
			</div>
		</div>

		<form method="dialog" class="modal-backdrop">
			<button @click.stop="closeChosenDialog" />
		</form>
	</dialog>
</template>