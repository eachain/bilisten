<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { CheckIcon, Cancel01Icon } from '@hugeicons/core-free-icons';
import { onActivated, ref } from 'vue';
import { getSettings, setSettings } from '../../libs/backend';
import { updateBufferTime } from '../../libs/player';

const bufferSeconds = ref(0);
let currentBuffer = bufferSeconds.value;

onActivated(() => {
	getSettings().then(({buffer}) => {
		bufferSeconds.value = buffer;
		currentBuffer = buffer;
	})
})

const showCommit = ref(false);
const modifyBuffer = () => {
	showCommit.value = bufferSeconds.value !== currentBuffer;
}

const commit = () => {
	if (bufferSeconds.value === currentBuffer) {
		return;
	}
	setSettings('buffer', `${bufferSeconds.value}`).then((buffer) => {
		bufferSeconds.value = Number(buffer);
		currentBuffer = bufferSeconds.value;
		showCommit.value = false;
		updateBufferTime(currentBuffer);
	})
};

const cancel = () => {
	bufferSeconds.value = currentBuffer;
	showCommit.value = false;
};
</script>

<template>
	<div class="flex items-center gap-2">
		<span>
			音频预加载
		</span>
		<div class="input w-20 input-xs px-2">
			<input
				type="number"
				class="validator"
				required
				placeholder=""
				min="1"
				v-model="bufferSeconds"
				@input="modifyBuffer"
				@keydown.enter="commit"
			/>
			<span class="label p-0 m-0 border-0">秒</span>
		</div>

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