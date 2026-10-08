<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { CheckIcon, Cancel01Icon, DragDropVerticalIcon } from '@hugeicons/core-free-icons';
import { VueDraggable } from 'vue-draggable-plus';
import { ref, onActivated } from 'vue';
import { getSettings, setSettings, SoundQualities } from '../../libs/backend';
import { updateQualityPrefer } from '../../libs/player';

const qualities = ref([...SoundQualities]);

const defaultPrefer = 'high,flac,dolby,mid,low';
let qualityPrefer = defaultPrefer;
let currentQuality = qualityPrefer;

const resort = () => {
	if (qualityPrefer === '') return;
	const mp = new Map();
	SoundQualities.forEach((sq) => mp.set(sq.q, sq));
	const qs = qualityPrefer.split(',');
	const results = [];
	for (const q of qs) {
		results.push(mp.get(q));
	}
	qualities.value = results;
}

onActivated(() => {
	resort();
	getSettings().then(({quality}) => {
		if (quality !== '') {
			qualityPrefer = quality;
			currentQuality = quality;
			resort();
		}
	})
})

const showCommit = ref(false);
const modifyQuality = () => {
	const qs: string[] = [];
	for (const q of qualities.value) {
		qs.push(q.q);
	}
	qualityPrefer = qs.join(',')
	showCommit.value = qualityPrefer !== currentQuality;
}

const commit = () => {
	if (qualityPrefer === currentQuality) {
		return;
	}
	setSettings('quality', qualityPrefer).then((quality) => {
		qualityPrefer = String(quality);
		currentQuality = qualityPrefer;
		showCommit.value = false;
		updateQualityPrefer(qualityPrefer);
		resort();
	})
};

const cancel = () => {
	qualityPrefer = currentQuality;
	showCommit.value = false;
	resort();
};
</script>

<template>
	<div class="flex items-start gap-2">
		<span>
			优先选择
		</span>
		<VueDraggable
			v-model="qualities"
			@update="modifyQuality"
			class="stats shadow stats-vertical"
		>
			<div v-for="sq in qualities" :key="sq.id" class="flex items-center gap-1 stat p-2 ps-1">
					<HugeiconsIcon :icon="DragDropVerticalIcon" class="opacity-50" :stroke-width="1" />
					{{ sq.name }}
				<div class="stat-desc">{{ sq.desc }}</div>
			</div>
		</VueDraggable>

		<div v-show="showCommit" class="flex gap-2">
			<button class="btn btn-outline btn-success btn-xs btn-circle" @click="commit">
				<HugeiconsIcon :icon="CheckIcon" :size="16" />
			</button>
			<button class="btn btn-outline btn-error btn-xs btn-circle" @click="cancel">
				<HugeiconsIcon :icon="Cancel01Icon" :size="16" />
			</button>
		</div>
	</div>
</template>