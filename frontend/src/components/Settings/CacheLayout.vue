<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue';
import { CheckIcon, Cancel01Icon } from '@hugeicons/core-free-icons';
import { onActivated, onDeactivated, ref } from 'vue';
import { clearCache, getCacheSize, getSettings, setSettings } from '../../libs/backend';

const { type, title, tooltip } = defineProps({
	type: String,
	title: String,
	tooltip: String,
});

const formatKNumber = (n: number) => {
	const s = (n / 1024).toFixed(1);
	if (s.endsWith('.0')) {
		return s.slice(0, s.length - 2);
	}
	return s;
}

const formatSize = (size: `${number}`) => {
	const n = Number(size);
	if (n < 1024) return `${size}B`;
	if (n < 1024 * 1024) return `${formatKNumber(n)}KB`;
	if (n < 1024 * 1024 * 1024) return `${formatKNumber(n / 1024)}MB`;
	return `${formatKNumber(n / 1024 / 1024)}GB`;
}

const cacheSizeUsed = ref('');
const cacheSizeLimit = ref(0);
let currentLimit = cacheSizeLimit.value;

let intv: any = undefined;
onActivated(() => {
	if (type !== 'audio' && type !== 'image') {
		return;
	}

	getSettings().then(({ audio, image }) => {
		if (type === 'audio') {
			cacheSizeLimit.value = audio;
		} else if (type === 'image') {
			cacheSizeLimit.value = image;
		}
		currentLimit = cacheSizeLimit.value;
	})

	if (intv) clearInterval(intv);
	intv = setInterval(() => {
		getCacheSize(type).then(size => {
			cacheSizeUsed.value = formatSize(size);
		})
	}, 1000)
})
onDeactivated(() => {
	if (intv) clearInterval(intv);
})

const clearEnabled = ref(true);
const clear = () => {
	if (type === 'audio' || type === 'image') {
		clearEnabled.value = false;
		clearCache(type).then(() => clearEnabled.value = true);
	}
}

const showCommit = ref(false);
const modifyLimit = () => {
	showCommit.value = cacheSizeLimit.value !== currentLimit;
}

const commit = () => {
	if (cacheSizeLimit.value === currentLimit) {
		return;
	}
	if (type !== 'audio' && type !== 'image') {
		return;
	}
	setSettings(type, `${cacheSizeLimit.value}`).then((limit) => {
		cacheSizeLimit.value = Number(limit);
		currentLimit = cacheSizeLimit.value;
		showCommit.value = false;
	})
};

const cancel = () => {
	cacheSizeLimit.value = currentLimit;
	showCommit.value = false;
};
</script>

<template>
	<div class="flex flex-col gap-2">
		<div class="flex items-center gap-2">
			<span class="tooltip tooltip-start" :data-tip="tooltip">
				{{ title }}
			</span>
			<div class="input w-24 input-xs px-2">
				<input type="number" class="validator" required placeholder="" min="0" v-model="cacheSizeLimit"
					@input="modifyLimit" @keydown.enter="commit" />
				<span class="label p-0 m-0 border-0">MB</span>
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

		<div class="text-xs tabular-nums ps-4 flex items-center gap-4">
			<span class="opacity-60">当前缓存：{{ cacheSizeUsed }}</span>
			<span class="btn btn-xs btn-error opacity-80" :class="clearEnabled ? '' : 'btn-disabled'" @click="clear">
				清空
			</span>
		</div>
	</div>
</template>