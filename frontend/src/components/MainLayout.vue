<script setup lang="ts">
import { onActivated } from 'vue';
// import { scrollToTop as _scrollToTop } from '../libs/scroll';

const {title, scrollToTop } = defineProps({
	title: String,
	scrollToTop: Boolean,
});

onActivated(async () => {
	if (scrollToTop) {
		(document.getElementById('app-main')! as HTMLElement).scrollTo({top: 0});
	}
})
</script>

<template>
	<div class="w-full flex flex-row justify-between items-center mb-4">
		<h2 v-if="title" class="text-2xl line-clamp-1">{{ title }}</h2>
		<slot name="header"></slot>
	</div>
	<slot></slot>
</template>