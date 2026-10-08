<script setup lang="ts">
import CacheLayout from './CacheLayout.vue';
</script>

<template>
	<div class="flex flex-col gap-4">
		<CacheLayout
			type="audio"
			title="音频缓存上限"
			tooltip="4分钟各音质音频大致参考：流畅2MB，标准4MB，高清8MB，杜比40MB，无损50MB"
		/>

		<CacheLayout
			type="image"
			title="图像缓存上限"
			tooltip="bilisten仅用缩略图，1MB大致可缓存200张图"
		/>
	</div>
</template>